# Raspberry Pi - Administration et maintenance

Le RPi `novoceo-os` héberge Zigbee2MQTT et fait le pont entre les appareils Zigbee et le broker MQTT sur k3s. Ce document couvre les scripts de maintenance installés sur le RPi.

## Sommaire

- [Sauvegarde de la carte SD](#sauvegarde-de-la-carte-sd)
- [Watchdog réseau](#watchdog-réseau)

---

## Sauvegarde de la carte SD

### Principe

La sauvegarde se fait **depuis le laptop** avec la carte SD branchée via un lecteur USB ou le slot intégré. Le script utilise `dd` pour créer une image brute compressée avec `gzip`. Les 3 dernières sauvegardes sont conservées automatiquement.

### Fichiers

| Fichier | Rôle |
|---------|------|
| `rpi/backup-sd.sh` | Script principal |
| `rpi/backups/` | Répertoire des images (exclu de git) |

### Utilisation

```bash
# 1. Brancher la carte SD sur le laptop

# 2. Identifier le périphérique
sudo ./rpi/backup-sd.sh detect

# Exemple de sortie :
# /dev/sdb  32G  [amovible] [USB] Generic USB Storage
# /dev/sda  512G            Samsung SSD

# 3. Sauvegarder
sudo ./rpi/backup-sd.sh backup /dev/sdb

# 4. Lister les sauvegardes existantes
sudo ./rpi/backup-sd.sh list

# 5. Restaurer
sudo ./rpi/backup-sd.sh restore /dev/sdb
```

### Fonctionnement interne

**Sauvegarde** : le script démonte automatiquement les partitions si elles ont été montées par le système, puis exécute :

```
dd if=/dev/sdb bs=4M | pv -s <taille> | gzip -c > rpi_sd_YYYYMMDD_HHMMSS.img.gz
```

- `dd bs=4M` : lit par blocs de 4 Mo pour des performances optimales
- `pv` : affiche la progression (vitesse, ETA, octets transférés) - nécessite le paquet `pv`
- `gzip` : compresse à la volée, réduit la taille de moitié environ

Si `pv` n'est pas installé, le script bascule sur `dd status=progress`.

**Rotation** : après chaque sauvegarde réussie, les images excédentaires les plus anciennes sont supprimées pour ne conserver que les 3 dernières.

**Restauration** : opération inverse - sélection interactive de la sauvegarde, confirmation explicite, puis :

```
pv <fichier.img.gz> | gunzip -c | dd of=/dev/sdb bs=4M
```

### Notes importantes

- Ne jamais sauvegarder une carte SD montée en écriture active - toujours la démonter depuis le RPi avant de l'extraire
- L'image contient la totalité du disque (partitions boot + root) - elle est restaurable sur n'importe quelle carte SD de taille identique ou supérieure
- Installer `pv` sur le laptop pour la barre de progression : `nix-env -iA nixpkgs.pv`

---

## Watchdog réseau

### Principe

Le RPi peut perdre sa connexion réseau sans s'en apercevoir (bug Zigbee2MQTT, pile IP gelée, etc.). Un watchdog vérifie la connectivité chaque minute et force un reboot si la perte de paquets dépasse 50%.

### Fichiers

| Fichier | Rôle |
|---------|------|
| `rpi/watchdog-net.sh` | Script de vérification |
| `rpi/watchdog-net.service` | Unité systemd (exécution du script) |
| `rpi/watchdog-net.timer` | Unité systemd (déclenchement toutes les minutes) |
| `rpi/watchdog-net.logrotate` | Rotation des logs (non utilisé avec systemd) |

### Architecture systemd : service + timer

Plutôt qu'une entrée cron, on utilise un couple **timer + service** systemd :

- Le **timer** (`watchdog-net.timer`) se déclenche toutes les minutes et appelle le service
- Le **service** (`watchdog-net.service`) exécute le script une fois (type `oneshot`) et se termine

Avantages sur cron :
- Logs centralisés dans journald (pas de fichier à gérer)
- Restart policy configurable
- Isolation des capabilities possible
- `systemctl status` donne un état immédiat

### Sécurité

Le script doit tourner en root car `reboot -f` l'exige. Pour limiter la surface d'attaque, le service est restreint dans `watchdog-net.service` :

```ini
CapabilityBoundingSet=CAP_SYS_BOOT CAP_NET_RAW
NoNewPrivileges=yes
```

- `CAP_SYS_BOOT` : autorise uniquement `reboot`
- `CAP_NET_RAW` : autorise uniquement les sockets raw (nécessaire pour `ping`)
- `NoNewPrivileges` : interdit toute élévation de privilèges ultérieure

Le script lui-même est protégé en lecture/exécution root uniquement :

```bash
chown root:root /opt/novoceo/watchdog-net.sh
chmod 700 /opt/novoceo/watchdog-net.sh
```

### Installation sur le RPi

```bash
# Depuis le laptop, copier les fichiers
scp rpi/watchdog-net.sh rpi/watchdog-net.service rpi/watchdog-net.timer pi@<IP_RPI>:/tmp/

# Sur le RPi
sudo mkdir -p /opt/novoceo
sudo mv /tmp/watchdog-net.sh /opt/novoceo/
sudo chown root:root /opt/novoceo/watchdog-net.sh
sudo chmod 700 /opt/novoceo/watchdog-net.sh

sudo mv /tmp/watchdog-net.service /tmp/watchdog-net.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now watchdog-net.timer
```

### Problème SELinux rencontré

Sur `novoceo-os` (openSUSE), SELinux est actif. Lors du `mv` depuis `/tmp/`, le fichier hérite du contexte SELinux `user_tmp_t` (contexte des fichiers temporaires). Systemd refuse d'exécuter un fichier avec ce label.

**Symptôme** :
```
watchdog-net.service: Failed at step EXEC spawning /opt/novoceo/watchdog-net.sh: Permission denied
```

**Diagnostic** :
```bash
ls -laZ /opt/novoceo/watchdog-net.sh
# -rwx------. 1 root root unconfined_u:object_r:user_tmp_t:s0 ...
#                                                  ^^^^^^^^^^^
#                                                  mauvais contexte
```

**Correction** : relabeler le fichier avec le contexte `bin_t` (exécutable système) :
```bash
sudo chcon -t bin_t /opt/novoceo/watchdog-net.sh
```

A retenir : toujours relabeler les fichiers copiés depuis `/tmp/` vers `/opt/` sur un système SELinux.

### Notifications MQTT

A chaque exécution, le script publie un heartbeat sur le broker MQTT central :

| Event | Condition | Payload |
|-------|-----------|---------|
| `heartbeat` | Toutes les minutes | `{"event":"heartbeat","loss":N}` |
| `reboot` | Avant reboot (perte > 50%) | `{"event":"reboot","loss":N}` |

- Broker : `192.168.1.128:32500`
- Topic : `rpi/watchdog-net`
- QoS : 0 (best-effort, le `|| true` garantit que le script ne bloque pas si le broker est injoignable)

Le recorder souscrit à `rpi/#` et persiste ces messages dans PostgreSQL comme les états Zigbee.

Pour surveiller en temps réel depuis le laptop :

```bash
mosquitto_sub -h 192.168.1.128 -p 32500 -t "rpi/watchdog-net" -v
```

### Vérification et logs

```bash
# Etat du timer (prochaine exécution)
systemctl status watchdog-net.timer

# Logs en temps réel
sudo journalctl -u watchdog-net.service -f

# Logs du jour
sudo journalctl -u watchdog-net.service --since today

# Exemple de log normal :
# May 20 09:58:01 novoceo-os watchdog-net.sh[1234]: perte=0%

# Exemple de reboot déclenché :
# May 20 09:58:01 novoceo-os watchdog-net.sh[1234]: perte=60%
# May 20 09:58:01 novoceo-os watchdog-net.sh[1234]: REBOOT déclenché (perte=60% > 50%)
```
