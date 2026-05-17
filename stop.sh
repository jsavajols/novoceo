#!/usr/bin/env bash
PIDS="$(cd "$(dirname "$0")" && pwd)/run/pids"

# Mode Docker
docker stop mosquitto 2>/dev/null && docker rm mosquitto 2>/dev/null && echo "[-] Mosquitto arrêté (Docker)" || true

# Mode natif
if [[ -f "$PIDS/mosquitto.pid" ]]; then
    pid=$(cat "$PIDS/mosquitto.pid")
    kill "$pid" 2>/dev/null && echo "[-] Mosquitto arrêté (natif, PID $pid)"
    rm -f "$PIDS/mosquitto.pid"
else
    pkill -f "mosquitto -c" 2>/dev/null && echo "[-] Mosquitto arrêté (natif, pkill)" || true
fi
