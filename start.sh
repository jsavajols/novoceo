#!/usr/bin/env bash
set -euo pipefail

SERVER_DIR="$(cd "$(dirname "$0")" && pwd)"
DATA_DIR="$SERVER_DIR/run"
PIDS="$DATA_DIR/pids"
MODE="${1:-native}"

# Arrête les instances précédentes
bash "$SERVER_DIR/stop.sh" 2>/dev/null || true

if [[ "$MODE" == "docker" ]]; then
    echo "[+] Démarrage de Mosquitto (Docker)..."
    docker rm -f mosquitto 2>/dev/null || true
    docker run -d --name mosquitto \
      -p 1883:1883 \
      -v "$DATA_DIR/mosquitto/mosquitto-docker.conf:/mosquitto/config/mosquitto.conf:ro" \
      -v "$DATA_DIR/mosquitto/data:/mosquitto/data" \
      eclipse-mosquitto:2
    echo "[+] Mosquitto actif sur port 1883 (Docker)"
else
    command -v mosquitto &>/dev/null || { echo "[x] mosquitto introuvable - lance 'nix-shell' d'abord"; exit 1; }
    mkdir -p "$PIDS"
    echo "[+] Démarrage de Mosquitto (natif)..."
    mosquitto -c "$DATA_DIR/mosquitto/mosquitto.conf" -d
    sleep 1
    MOSQ_PID=$(pgrep -f "mosquitto -c $DATA_DIR" || true)
    if [[ -z "$MOSQ_PID" ]]; then
        echo "[x] Mosquitto n'a pas démarré"
        exit 1
    fi
    echo "$MOSQ_PID" > "$PIDS/mosquitto.pid"
    echo "[+] Mosquitto actif sur port 1883 (natif, PID $MOSQ_PID)"
fi

echo ""
echo "  Lancer le monitor :  go run ./monitor/ tcp://100.64.0.1:1883"
echo "  Stopper :            ./stop.sh"
echo ""
