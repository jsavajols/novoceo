# Raspberry Pi - Installation et configuration de l'OS

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
