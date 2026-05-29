#!/bin/sh
# watchdog-net.sh - Reboot si la connectivité réseau est dégradée
# Planifié via cron toutes les minutes (watchdog-net.crontab)

PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

TARGET="8.8.8.8"
COUNT=10
THRESHOLD=50  # % d'erreurs au-delà duquel on reboot
REBOOT_MIN_TICKS=10  # nombre minimum de passages avant d'autoriser un reboot

MQTT_HOST="192.168.1.128"
MQTT_PORT="32500"
MQTT_TOPIC="rpi0/watchdog-net"

# Compteur persistant dans /run/ (tmpfs : remis à zéro au boot)
TICK_FILE="/run/watchdog-net-ticks"
ticks=$(cat "$TICK_FILE" 2>/dev/null)
ticks=$(( ${ticks:-0} + 1 ))
echo "$ticks" > "$TICK_FILE"

lost=$(ping -c "$COUNT" -q "$TARGET" 2>/dev/null \
    | grep -oE '[0-9]+%' | tr -d '%')

if [ -z "$lost" ]; then
    echo "ERREUR: ping injoignable ou sortie inattendue"
    lost=100
fi

echo "perte=${lost}% ticks=${ticks}"

mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC" \
    -m "{\"event\":\"heartbeat\",\"loss\":${lost},\"ticks\":${ticks}}" 2>/dev/null || true

if [ "$lost" -gt "$THRESHOLD" ] && [ "$ticks" -ge "$REBOOT_MIN_TICKS" ]; then
    echo "REBOOT déclenché (perte=${lost}% > ${THRESHOLD}%, ticks=${ticks})"
    mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC" \
        -m "{\"event\":\"reboot\",\"loss\":${lost},\"ticks\":${ticks}}" 2>/dev/null || true
    sleep 1
    reboot -f
elif [ "$lost" -gt "$THRESHOLD" ]; then
    echo "Perte élevée mais reboot différé (ticks=${ticks} < ${REBOOT_MIN_TICKS})"
fi
