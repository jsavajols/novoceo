# Raspberry Pi - Installation et configuration de l'OS

Deux configurations OS sont utilisées selon le matériel :

- **RPi3** (`novoceo-os`) : openSUSE MicroOS avec systemd
- **RPi Zero 2W** : Alpine Linux avec OpenRC

---

## RPi Zero 2W - Alpine Linux

Alpine Linux est une distribution minimaliste (~130 Mo installé) bien adaptée au RPi Zero 2W dont les ressources sont limitées (512 Mo RAM, CPU ARM Cortex-A53 quad-core 1 GHz).

### Specificites

- Init system : **OpenRC** (pas systemd)
- Gestionnaire de paquets : `apk`
- Privilege escalation : `doas` (pas `sudo` par defaut)
- Pas de journald : les logs vont dans des fichiers sous `/var/log/`

### Configuration de doas

`doas` est l'outil de privilege escalation par défaut sur Alpine (équivalent `sudo`). Il nécessite une configuration explicite - par défaut aucun utilisateur ne peut élever ses privilèges.

```bash
# En root (su -)

# Verifier que l'utilisateur est dans le groupe wheel
id jerome

# Si absent
adduser jerome wheel

# Activer la regle wheel dans /etc/doas.conf
sed -i 's/^# permit persist :wheel/permit persist :wheel/' /etc/doas.conf
```

Le fichier `/etc/doas.conf` est lisible uniquement par root (mode 400) - c'est normal et attendu.

### Packages necessaires

```bash
# Zigbee2MQTT (installe le binaire dans /usr/bin/zigbee2mqtt)
doas apk add zigbee2mqtt

# mosquitto-clients pour le watchdog MQTT
doas apk add mosquitto-clients
```

Node.js n'est pas a installer séparément - le package `zigbee2mqtt` d'Alpine embarque ses propres dépendances.

### Configuration de Zigbee2MQTT

Le fichier de configuration attendu par le service est `/root/.z2m/configuration.yaml`. C'est le répertoire de données par défaut du package Alpine (`ZIGBEE2MQTT_DATA=/root/.z2m`).

Si ce répertoire est absent ou vide, zigbee2mqtt démarre en mode onboarding (interface web sur le port 8080) au lieu de se connecter au réseau Zigbee.

### Persistance des services

Sur Alpine Linux, les services activés par `rc-update` sont persistés dans `/etc/runlevels/`. Contrairement a un systeme immuable, les modifications sont directement appliquées sur le système de fichiers.

---

## RPi3 - openSUSE MicroOS

`novoceo-os` tourne sous **openSUSE MicroOS**, un OS immuable basé sur openSUSE Tumbleweed. L'immutabilité implique que les modifications du système de fichiers racine passent par `transactional-update`, qui crée un nouveau snapshot btrfs activé au prochain reboot.

## Packages complémentaires

### mosquitto-clients

Requis par `watchdog-net.sh` pour publier les heartbeats MQTT vers le broker central.

```bash
# Rechercher le bon package (les clients sont séparés du broker)
zypper search mosquitto
# S  | Name              | Summary
# ---+-------------------+------------------
#    | mosquitto-clients | Client for Mosquitto   <-- celui-ci
#  i | mosquitto         | A MQTT v3.1/v3.1.1 Broker

# Installer (transactional-update crée un nouveau snapshot)
sudo transactional-update pkg install mosquitto-clients

# Rebooter pour activer le snapshot
sudo reboot
```

> Ne pas installer `mosquitto` (le broker) : il n'est pas nécessaire sur le RPi, le broker tourne dans le cluster k3s.
