# Développement local

## Prérequis

- Go 1.25+
- Mosquitto (broker MQTT)
- PostgreSQL (optionnel pour recorder/api)
- Nix (optionnel, shell.nix fourni)
- Docker buildx avec QEMU (pour les builds multi-arch)

## RPi — Zigbee2MQTT

Zigbee2MQTT tourne en service systemd sur le RPi. Les fichiers sont dans `rpi/` :

| Fichier | Description |
|---------|-------------|
| `zigbee2mqtt.service` | Unité systemd — à copier dans `/etc/systemd/system/` |
| `start-z2m·sh` | Script de lancement manuel (debug, logs dans le terminal) |

```bash
# Installer le service
sudo cp rpi/zigbee2mqtt.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now zigbee2mqtt

# Logs
journalctl -u zigbee2mqtt -f
```

```bash
nix-shell   # charge l'environnement complet
```

## Fichier .env

Copier et adapter pour le dev local :

```env
MQTT_BROKER=tcp://192.168.1.128:32500
API_TOKEN=<token>
DB_HOST=51.159.109.144
DB_PORT=30015
DB_USER=rpi-novoceo
DB_PASSWORD=<password>
DB_NAME=novoceo
```

Ce fichier est chargé automatiquement par tous les services via `godotenv.Load()`.

## Broker mosquitto local

```bash
# Mode natif (nix-shell requis)
./start.sh

# Mode Docker
./start.sh docker

# Arrêter
./stop.sh

# Logs Docker
docker logs -f mosquitto
```

Le broker écoute sur `0.0.0.0:1883`.

> Docker rootless : les autres conteneurs (snug, recorder) doivent utiliser
> l'IP VPN `100.64.0.1` et non `localhost`.

## monitor — dashboard terminal

Affiche en temps réel les devices Zigbee et leurs valeurs (lecture seule).

```bash
# Avec le broker k3s
go run ./monitor/ tcp://192.168.1.128:32500

# Avec un broker local
go run ./monitor/ tcp://localhost:1883

# Binaire compilé
go build -o monitor/monitor ./monitor/ && ./monitor/monitor tcp://192.168.1.128:32500
```

## snug — lier un bouton à une prise

```bash
go build -o snug/snug ./snug/

# Flags par défaut (broker k3s)
./snug/snug -plug Prise -action single

# Tous les flags
./snug/snug \
  -broker tcp://192.168.1.128:32500 \
  -button Bouton \
  -plug Prise \
  -action single \
  -rpi novoceo-os
```

### Docker / Registry

```bash
# Build local
docker build -f snug/Dockerfile -t snug .
docker run -d --name snug --restart unless-stopped \
  snug:latest -broker tcp://100.64.0.1:1883 -plug Prise -action single
docker logs -f snug
docker stop snug && docker rm snug

# Build multi-arch + push (amd64 + arm64)
./k8s/snug/dev-build-deploy.sh
```

## send-device — commande ponctuelle

Connexion one-shot : publie et quitte immédiatement.

```bash
go build -o send-device/send-device ./send-device/

# Commandes simples
./send-device/send-device -device Prise -state TOGGLE
./send-device/send-device -device Prise -state ON
./send-device/send-device -device Prise -state OFF

# Payload JSON brut
./send-device/send-device -device Prise -payload '{"state":"ON","brightness":128}'

# Topologie multi-RPi
./send-device/send-device -device Prise -rpi novoceo-os -state TOGGLE

# Broker custom
./send-device/send-device -broker tcp://localhost:1883 -device Prise -state TOGGLE
```

## recorder — persistance des events

```bash
# Local (charge .env automatiquement)
go run ./recorder/

# Binaire compilé
go build -o recorder/recorder ./recorder/ && ./recorder/recorder
```

### Docker

```bash
docker build -f recorder/Dockerfile -t recorder .
docker run -d --name recorder --restart unless-stopped \
  --env-file .env \
  recorder:latest
```

## api — service REST

```bash
go run ./api/
# Ecoute sur :5000 par défaut
```

## front — dashboard web

```bash
# Avec l'api en local sur :5000
API_URL=http://localhost:5000 API_TOKEN=<token> go run ./front/
# Ecoute sur :3000 par défaut
```

Ouvrir `http://localhost:3000` dans le navigateur.

## Build de tous les binaires

```bash
go build -o monitor/monitor   ./monitor/
go build -o snug/snug         ./snug/
go build -o send-device/send-device ./send-device/
go build -o recorder/recorder ./recorder/
```
