#!/bin/sh
# watchdog-net.sh - Reboot si la connectivité réseau est dégradée
# Planifié via cron toutes les minutes (watchdog-net.crontab)

PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

TARGET="8.8.8.8"
COUNT=10
THRESHOLD=50  # % d'erreurs au-delà duquel on reboot

MQTT_HOST="192.168.1.128"
MQTT_PORT="32500"
MQTT_TOPIC="rpi0/watchdog-net"

lost=$(ping -c "$COUNT" -q "$TARGET" 2>/dev/null \
    | grep -oE '[0-9]+%' | tr -d '%')

if [ -z "$lost" ]; then
    echo "ERREUR: ping injoignable ou sortie inattendue"
    lost=100
fi

echo "perte=${lost}%"

mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC" \
    -m "{\"event\":\"heartbeat\",\"loss\":${lost}}" 2>/dev/null || true

if [ "$lost" -gt "$THRESHOLD" ]; then
    echo "REBOOT déclenché (perte=${lost}% > ${THRESHOLD}%)"
    mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC" \
        -m "{\"event\":\"reboot\",\"loss\":${lost}}" 2>/dev/null || true
    sleep 1
    reboot -f
fi
