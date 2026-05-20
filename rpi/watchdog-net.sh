#!/usr/bin/env bash
# watchdog-net.sh - Reboot si la connectivité réseau est dégradée
# Géré par systemd : watchdog-net.timer + watchdog-net.service

TARGET="8.8.8.8"
COUNT=10
THRESHOLD=50  # % d'erreurs au-delà duquel on reboot

MQTT_HOST="192.168.1.128"
MQTT_PORT="32500"
MQTT_TOPIC="rpi/watchdog-net"

lost=$(ping -c "$COUNT" -q "$TARGET" 2>/dev/null \
    | awk '/packets transmitted/ { gsub(/%/,"",$6); print $6 }')

if [[ -z "$lost" ]]; then
    echo "ERREUR: ping injoignable ou sortie inattendue"
    lost=100
fi

echo "perte=${lost}%"

mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC" \
    -m "{\"event\":\"heartbeat\",\"loss\":${lost}}" 2>/dev/null || true

if (( lost > THRESHOLD )); then
    echo "REBOOT déclenché (perte=${lost}% > ${THRESHOLD}%)"
    mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC" \
        -m "{\"event\":\"reboot\",\"loss\":${lost}}" 2>/dev/null || true
    sleep 1
    reboot -f
fi
