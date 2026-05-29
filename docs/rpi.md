# Raspberry Pi - Administration et maintenance

Deux RPi sont utilisés dans le projet :

- `novoceo-os` (RPi 3) : héberge Zigbee2MQTT sous openSUSE MicroOS avec systemd
- RPi Zero 2W : variante légère sous Alpine Linux avec OpenRC

## Sommaire

- [Zigbee2MQTT sur RPi Zero 2W (Alpine Linux)](#zigbee2mqtt-sur-rpi-zero-2w-alpine-linux)
- [Watchdog réseau sur RPi Zero 2W (Alpine Linux)](#watchdog-réseau-sur-rpi-zero-2w-alpine-linux)
- [Configuration WiFi](#configuration-wifi---changer-de-point-dacces)
- [Sauvegarde de la carte SD](#sauvegarde-de-la-carte-sd)
- [Watchdog réseau et Z2M sur RPi3 (openSUSE MicroOS)](#watchdog-réseau-et-z2m-sur-rpi3-opensuse-micros)

---

## Zigbee2MQTT sur RPi Zero 2W (Alpine Linux)

### Contexte

Le RPi Zero 2W tourne sous Alpine Linux qui utilise **OpenRC** comme système d'init (pas systemd). Le service est défini dans `rpi/rpi0/zigbee2mqtt.initd`.

Zigbee2MQTT est installé via le gestionnaire de paquets Alpine (`apk`), le binaire est donc à `/usr/bin/zigbee2mqtt`. Le fichier de configuration se trouve dans `/root/.z2m/configuration.yaml` et le service tourne en root (nécessaire pour l'accès au port série du dongle Zigbee).

### Differences avec le RPi3

| | RPi3 (openSUSE MicroOS) | RPi Zero 2W (Alpine Linux) |
|---|---|---|
| Init system | systemd | OpenRC |
| Fichier service | `rpi/rpi3/zigbee2mqtt.service` | `rpi/rpi0/zigbee2mqtt.initd` |
| Installation z2m | manuelle (`/opt/zigbee2mqtt`) | `apk add zigbee2mqtt` |
| Config z2m | `ZIGBEE2MQTT_DATA` | `/root/.z2m/` |
| Restart on failure | `Restart=on-failure` | `supervisor=supervise-daemon` |
| Logs | journald | `/var/log/zigbee2mqtt/zigbee2mqtt.log` |
| Network dependency | `After=network-online.target` | `need net` |
| Privilege escalation | `sudo` | `doas` (groupe `wheel`) |

### Prérequis : configurer doas

Sur Alpine, `doas` requiert une configuration explicite. Vérifier que jerome est dans le groupe `wheel` et que la règle est activée :

```bash
# En root (su -)
id jerome   # verifier la presence de wheel dans les groupes
# Si absent : adduser jerome wheel

# Activer la regle wheel dans doas
sed -i 's/^# permit persist :wheel/permit persist :wheel/' /etc/doas.conf
```

### Installation de zigbee2mqtt

```bash
# En root (su -)
apk add zigbee2mqtt

# Le binaire est installe a /usr/bin/zigbee2mqtt
# La configuration doit etre creee dans /root/.z2m/configuration.yaml
```

### Deploiement du service

```bash
# Depuis le laptop, copier le fichier init
scp rpi/rpi0/zigbee2mqtt.initd jerome@100.64.0.11:/tmp/

# Sur le RPi Zero 2W
doas cp /tmp/zigbee2mqtt.initd /etc/init.d/zigbee2mqtt
doas chmod +x /etc/init.d/zigbee2mqtt

# Activer au demarrage et demarrer
doas rc-update add zigbee2mqtt default
doas rc-service zigbee2mqtt start
```

Le `chmod +x` est obligatoire - OpenRC refuse d'executer un script non executable.

### Gestion du service

```bash
# Etat
rc-service zigbee2mqtt status

# Demarrer / arreter / redemarrer
doas rc-service zigbee2mqtt start
doas rc-service zigbee2mqtt stop
doas rc-service zigbee2mqtt restart

# Logs en temps reel
tail -f /var/log/zigbee2mqtt/zigbee2mqtt.log

# Desactiver le demarrage automatique
doas rc-update del zigbee2mqtt default
```

### Supervision automatique

La surveillance de Z2M fonctionne sur trois niveaux indépendants et complémentaires :

**Niveau 1 - Watchdog interne Z2M** (`Z2M_WATCHDOG`)

Configuré dans `zigbee2mqtt.initd` via la variable `Z2M_WATCHDOG="1,1,2,5,10"`. Quand Z2M détecte lui-même une défaillance de son contrôleur Zigbee (dongle qui décroche, etc.), il redémarre le contrôleur en mémoire selon ces délais (en minutes). Le nombre de valeurs = nombre max de tentatives. Passé ce seuil, Z2M quitte proprement pour laisser le niveau suivant prendre la main.

**Niveau 2 - supervise-daemon** (OpenRC)

`supervisor=supervise-daemon` dans le script init demande a OpenRC de surveiller le process et de le relancer automatiquement si Z2M quitte, équivalent au `Restart=on-failure` de systemd. Prend le relais quand Z2M a épuisé ses retries internes.

**Niveau 3 - watchdog-net.sh** (cron, toutes les minutes)

Vérifie `rc-service zigbee2mqtt status` à chaque passage. Si Z2M est down malgré les deux niveaux précédents : jusqu'à 3 tentatives de `rc-service restart` (compteur persisté dans `/run/`, remis à zéro au boot). Si Z2M reste irrécupérable : `reboot -f` système.

Tous les événements sont publiés sur `rpi0/watchdog-z2m` en MQTT (voir section watchdog réseau).

### Points d'attention

- Le service doit tourner en **root** pour accéder au port série du dongle Zigbee - ne pas ajouter `command_user`
- `ZIGBEE2MQTT_DATA` doit pointer vers `/root/.z2m` - z2m entre en mode onboarding si la config est introuvable
- Les logs vont dans `/var/log/zigbee2mqtt/` via `supervise_daemon_args --stdout/--stderr` - les variables `output_log`/`error_log` ne sont pas supportées par `supervise-daemon` sur Alpine

---

## Watchdog réseau et Z2M sur RPi Zero 2W (Alpine Linux)

### Principe

Vérification de la connectivité réseau et de l'état de Zigbee2MQTT toutes les minutes. Le script reboot si la perte de paquets dépasse 50%, et surveille Z2M en relayant automatiquement le service avant de rebooter en dernier recours. Sur Alpine, le timer systemd est remplacé par **crond** (busybox).

### Fichiers

| Fichier | Rôle |
|---------|------|
| `rpi/rpi0/watchdog-net.sh` | Script de vérification |
| `rpi/rpi0/watchdog-net.crontab` | Entrée cron (remplace le timer systemd) |
| `rpi/rpi0/watchdog-net.logrotate` | Rotation des logs |
| `rpi/rpi0/watchdog-toggle.sh` | Suspend/reactive le watchdog |

### Differences avec le RPi3

| | RPi3 (openSUSE MicroOS) | RPi Zero 2W (Alpine Linux) |
|---|---|---|
| Planification | systemd timer | crond (busybox) |
| Fichiers | `.service` + `.timer` | `.crontab` |
| Shell | bash | sh (busybox ash) |
| Logs | journald | `/var/log/watchdog-net.log` |

Le script est réécrit en `sh` POSIX (pas de `[[ ]]` ni `(( ))`) car Alpine utilise busybox ash par défaut.

### Installation sur le RPi Zero 2W

```bash
# Depuis le laptop
scp rpi/rpi0/watchdog-net.sh rpi/rpi0/watchdog-net.crontab rpi/rpi0/watchdog-net.logrotate jerome@100.64.0.11:/tmp/

# Sur le RPi Zero 2W (en root)
mkdir -p /opt/novoceo
cp /tmp/watchdog-net.sh /opt/novoceo/
chown root:root /opt/novoceo/watchdog-net.sh
chmod 700 /opt/novoceo/watchdog-net.sh

# Installer l'entree cron (la redirection >> doit passer par sh -c, pas doas directement)
doas sh -c 'cat /tmp/watchdog-net.crontab >> /etc/crontabs/root'

# Installer la rotation des logs
cp /tmp/watchdog-net.logrotate /etc/logrotate.d/watchdog-net

# Activer et démarrer crond si pas déjà fait
rc-update add crond default
rc-service crond start

# Redemarrer crond pour qu'il relise /etc/crontabs/root
rc-service crond restart
```

> crond busybox ne relit pas le crontab à chaud - tout ajout ou modification nécessite un `rc-service crond restart`.

### Vérification

```bash
# Verifier que crond tourne
rc-service crond status

# Voir le crontab root
crontab -l

# Logs en temps réel (après la première minute)
tail -f /var/log/watchdog-net.log

# Exemple de sortie normale
# perte=0% ticks=5
# z2m=ok

# Exemple reboot différé (ticks < 10, protection au boot)
# perte=80% ticks=3
# Perte élevée mais reboot différé (ticks=3 < 10)

# Exemple de reboot réseau déclenché
# perte=60% ticks=12
# REBOOT déclenché (perte=60% > 50%, ticks=12)

# Exemple Z2M down avec restart automatique
# z2m=down
# Redémarrage z2m (tentative 1/3)

# Exemple reboot suite à Z2M irrécupérable
# z2m=down
# z2m non récupéré après 3 tentatives - reboot système
```

### Gestion manuelle du watchdog

Pour suspendre temporairement le watchdog (maintenance, changement reseau) :

```bash
# Depuis le laptop
scp rpi/rpi0/watchdog-toggle.sh jerome@100.64.0.11:/tmp/
doas cp /tmp/watchdog-toggle.sh /opt/novoceo/
doas chmod 700 /opt/novoceo/watchdog-toggle.sh

# Usage (en root)
doas /opt/novoceo/watchdog-toggle.sh suspend   # stoppe le watchdog
doas /opt/novoceo/watchdog-toggle.sh resume    # remet en route
doas /opt/novoceo/watchdog-toggle.sh status    # etat actuel
```

Le script commente/decommente la ligne cron dans `/etc/crontabs/root` et redémarre crond.

### Notifications MQTT

**Topic `rpi0/watchdog-net`** - connectivité réseau :

| Event | Condition | Payload |
|-------|-----------|---------|
| `heartbeat` | Toutes les minutes | `{"event":"heartbeat","loss":N,"ticks":N}` |
| `reboot` | Avant reboot réseau (perte > 50%, ticks >= 10) | `{"event":"reboot","loss":N,"ticks":N}` |

**Topic `rpi0/watchdog-z2m`** - état Zigbee2MQTT :

| Event | Condition | Payload |
|-------|-----------|---------|
| `heartbeat` | Toutes les minutes si Z2M est ok | `{"event":"heartbeat","status":"ok","restarts":N}` |
| `restart` | Tentative de relance Z2M | `{"event":"restart","attempt":N,"max":3}` |
| `reboot` | Reboot après 3 échecs Z2M | `{"event":"reboot","reason":"z2m_unrecoverable","restarts":3}` |

- Broker : `192.168.1.128:32500`
- Topic réseau : `rpi0/watchdog-net` (differe du RPi3 qui utilise `rpi/watchdog-net`)
- Topic Z2M : `rpi0/watchdog-z2m`
- `ticks` : compteur de passages depuis le boot (tmpfs `/run/`, remis à zéro au reboot) - le reboot réseau ne se déclenche qu'à partir de ticks=10 pour éviter les faux positifs au démarrage

```bash
# Surveiller depuis le laptop
mosquitto_sub -h 192.168.1.128 -p 32500 -t "rpi0/watchdog-net" -v
mosquitto_sub -h 192.168.1.128 -p 32500 -t "rpi0/watchdog-z2m" -v
mosquitto_sub -h 192.168.1.128 -p 32500 -t "rpi0/#" -v
```

---

## Configuration WiFi - changer de point d'acces

### RPi Zero 2W (Alpine Linux)

Alpine utilise `wpa_supplicant` + `udhcpc`. Le fichier de configuration est `/etc/wpa_supplicant/wpa_supplicant.conf`.

```bash
# Generer un bloc network avec psk hashe (recommande)
wpa_passphrase "NouveauSSID" "MotDePasse"
# Copier le bloc network{...} dans le fichier de config

doas vi /etc/wpa_supplicant/wpa_supplicant.conf
```

Format attendu :

```
network={
    ssid="NouveauSSID"
    #psk="MotDePasse"
    psk=<hash_genere_par_wpa_passphrase>
}
```

Appliquer sans reboot :

```bash
doas ifdown wlan0 && doas ifup wlan0
# ou
doas rc-service networking restart
```

Verifier la connexion :

```bash
ip addr show wlan0
iwconfig wlan0
```

### RPi3 (openSUSE MicroOS avec NetworkManager)

```bash
# Voir les reseaux disponibles
nmcli device wifi list

# Se connecter a un nouveau reseau (cree et active la connexion)
sudo nmcli device wifi connect "NouveauSSID" password "MotDePasse"

# Modifier une connexion existante
nmcli connection show                                              # lister
sudo nmcli connection modify "NomConnexion" wifi-sec.psk "NouveauMDP"
sudo nmcli connection up "NomConnexion"
```

NetworkManager ne necessite pas de reboot ni de `transactional-update` pour changer de reseau WiFi.

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

## Watchdog réseau et Z2M sur RPi3 (openSUSE MicroOS)

### Principe

Vérification de la connectivité réseau et de l'état de Zigbee2MQTT toutes les minutes via systemd timer. Le script reboot si la perte de paquets dépasse 50%, et surveille Z2M en le relançant via `systemctl` avant de rebooter en dernier recours.

### Supervision Zigbee2MQTT

Trois niveaux complémentaires, identiques au RPi0 mais adaptés à systemd :

**Niveau 1 - Watchdog interne Z2M** (`Z2M_WATCHDOG=1,1,2,5,10`)
Configuré dans `rpi/rpi3/zigbee2mqtt.service`. Z2M redémarre son contrôleur Zigbee en mémoire selon ces délais (minutes) avant de quitter. Visible dans journald : `Starting Zigbee2MQTT with watchdog (60000,60000,120000,300000,600000)`.

**Niveau 2 - systemd** (`Restart=on-failure`)
systemd relance le process Z2M si celui-ci quitte avec une erreur.

**Niveau 3 - watchdog-net.sh** (timer, toutes les minutes)
Vérifie `systemctl is-active zigbee2mqtt`. Si Z2M est down : jusqu'à 3 `systemctl restart` (compteur dans `/run/`, remis à zéro au boot), puis `reboot -f` système.

> **Note** : `Type=notify` + `WatchdogSec` (watchdog natif systemd) a été testé mais ne fonctionne pas : le module natif `unix-dgram` nécessaire pour `sd_notify` n'est pas compilé pour Node.js 24 sur ARM64. Le service reste en `Type=simple`.

### Fichiers

| Fichier | Rôle |
|---------|------|
| `rpi/rpi3/watchdog-net.sh` | Script de vérification réseau et Z2M |
| `rpi/rpi3/watchdog-net.service` | Unité systemd (exécution du script) |
| `rpi/rpi3/watchdog-net.timer` | Unité systemd (déclenchement toutes les minutes) |
| `rpi/rpi3/watchdog-net.logrotate` | Rotation des logs (non utilisé avec systemd) |
| `rpi/rpi3/watchdog-toggle.sh` | Suspend/reactive le watchdog |

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

`mosquitto-clients` doit être installé sur le RPi avant de déployer le script (voir [docs/rpi-os.md](rpi-os.md)).

```bash
# Depuis le laptop, copier les fichiers
scp rpi/watchdog-net.sh rpi/watchdog-net.service rpi/watchdog-net.timer jerome@100.64.0.8:/tmp/

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

**Topic `rpi/watchdog-net`** - connectivité réseau :

| Event | Condition | Payload |
|-------|-----------|---------|
| `heartbeat` | Toutes les minutes | `{"event":"heartbeat","loss":N,"ticks":N}` |
| `reboot` | Avant reboot réseau (perte > 50%, ticks >= 10) | `{"event":"reboot","loss":N,"ticks":N}` |

**Topic `rpi/watchdog-z2m`** - état Zigbee2MQTT :

| Event | Condition | Payload |
|-------|-----------|---------|
| `heartbeat` | Toutes les minutes si Z2M est ok | `{"event":"heartbeat","status":"ok","restarts":N}` |
| `restart` | Tentative de relance Z2M | `{"event":"restart","attempt":N,"max":3}` |
| `reboot` | Reboot après 3 échecs Z2M | `{"event":"reboot","reason":"z2m_unrecoverable","restarts":3}` |

- Broker : `192.168.1.128:32500`
- QoS : 0 (best-effort, le `|| true` garantit que le script ne bloque pas si le broker est injoignable)
- `ticks` : compteur de passages depuis le boot (tmpfs `/run/`, remis à zéro au reboot) - le reboot réseau ne se déclenche qu'à partir de ticks=10

Le recorder souscrit à `rpi/#` et persiste ces messages dans PostgreSQL comme les états Zigbee.

```bash
# Surveiller depuis le laptop
mosquitto_sub -h 192.168.1.128 -p 32500 -t "rpi/watchdog-net" -v
mosquitto_sub -h 192.168.1.128 -p 32500 -t "rpi/watchdog-z2m" -v
mosquitto_sub -h 192.168.1.128 -p 32500 -t "rpi/#" -v
```

### Test du watchdog

Simuler une perte de connectivité sans iptables (route blackhole) :

```bash
# Bloquer 8.8.8.8
sudo ip route add blackhole 8.8.8.8/32

# Débloquer
sudo ip route del blackhole 8.8.8.8/32
```

Surveiller le témoin MQTT pendant le test :

```bash
mosquitto_sub -h 192.168.1.128 -p 32500 -t "rpi/watchdog-net" -v
```

### Gestion manuelle du watchdog

```bash
# Depuis le laptop
scp rpi/rpi3/watchdog-toggle.sh jerome@100.64.0.8:/tmp/
sudo mv /tmp/watchdog-toggle.sh /opt/novoceo/
sudo chmod 700 /opt/novoceo/watchdog-toggle.sh
sudo chcon -t bin_t /opt/novoceo/watchdog-toggle.sh  # SELinux

# Usage
sudo /opt/novoceo/watchdog-toggle.sh suspend   # stoppe le watchdog
sudo /opt/novoceo/watchdog-toggle.sh resume    # remet en route
sudo /opt/novoceo/watchdog-toggle.sh status    # etat actuel
```

La suspension utilise `systemctl stop/start watchdog-net.timer` et ne survit pas a un reboot.

### Vérification et logs

```bash
# Etat du timer (prochaine exécution)
systemctl status watchdog-net.timer

# Logs en temps réel
sudo journalctl -u watchdog-net.service -f

# Logs du jour
sudo journalctl -u watchdog-net.service --since today

# Exemple de log normal :
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: perte=0% ticks=5
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: z2m=ok

# Exemple reboot différé (protection au boot, ticks < 10) :
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: perte=80% ticks=3
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: Perte élevée mais reboot différé (ticks=3 < 10)

# Exemple de reboot réseau déclenché :
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: perte=60% ticks=12
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: REBOOT déclenché (perte=60% > 50%, ticks=12)

# Exemple Z2M down avec restart automatique :
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: z2m=down
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: Redémarrage z2m (tentative 1/3)

# Exemple reboot suite à Z2M irrécupérable :
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: z2m=down
# May 29 09:58:01 novoceo-os watchdog-net.sh[1234]: z2m non récupéré après 3 tentatives - reboot système
```
