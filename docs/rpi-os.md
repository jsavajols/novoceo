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

### Packages necessaires

```bash
# Node.js pour Zigbee2MQTT
doas apk add nodejs npm

# mosquitto-clients pour le watchdog MQTT
doas apk add mosquitto-clients
```

### Persistance des services

Sur Alpine Linux, les services actives par `rc-update` sont persistes dans `/etc/runlevels/`. Contrairement a un systeme immuable, les modifications sont directement appliquees sur le systeme de fichiers.

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
