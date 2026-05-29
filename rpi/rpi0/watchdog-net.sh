#!/bin/sh
# watchdog-net.sh - Surveillance réseau et Zigbee2MQTT
# Planifié via cron toutes les minutes (watchdog-net.crontab)

PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

TARGET="8.8.8.8"
COUNT=10
THRESHOLD=50  # % d'erreurs au-delà duquel on reboot
REBOOT_MIN_TICKS=10  # nombre minimum de passages avant d'autoriser un reboot

MQTT_HOST="100.64.0.10"
MQTT_PORT="32500"
MQTT_TOPIC="rpi0/watchdog-net"
MQTT_TOPIC_Z2M="rpi0/watchdog-z2m"
Z2M_SERVICE="zigbee2mqtt"
Z2M_RESTART_MAX=3
Z2M_RESTART_FILE="/run/watchdog-z2m-restarts"

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

# --- Surveillance Zigbee2MQTT ---
z2m_restarts=$(cat "$Z2M_RESTART_FILE" 2>/dev/null)
z2m_restarts=${z2m_restarts:-0}

if rc-service "$Z2M_SERVICE" status >/dev/null 2>&1; then
    echo "z2m=ok"
    mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC_Z2M" \
        -m "{\"event\":\"heartbeat\",\"status\":\"ok\",\"restarts\":${z2m_restarts}}" 2>/dev/null || true
else
    echo "z2m=down"
    if [ "$z2m_restarts" -lt "$Z2M_RESTART_MAX" ]; then
        z2m_restarts=$(( z2m_restarts + 1 ))
        echo "$z2m_restarts" > "$Z2M_RESTART_FILE"
        echo "Redémarrage z2m (tentative ${z2m_restarts}/${Z2M_RESTART_MAX})"
        mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC_Z2M" \
            -m "{\"event\":\"restart\",\"attempt\":${z2m_restarts},\"max\":${Z2M_RESTART_MAX}}" 2>/dev/null || true
        rc-service "$Z2M_SERVICE" restart >/dev/null 2>&1 || true
    else
        echo "z2m non récupéré après ${Z2M_RESTART_MAX} tentatives - reboot système"
        mosquitto_pub -h "$MQTT_HOST" -p "$MQTT_PORT" -t "$MQTT_TOPIC_Z2M" \
            -m "{\"event\":\"reboot\",\"reason\":\"z2m_unrecoverable\",\"restarts\":${z2m_restarts}}" 2>/dev/null || true
        sleep 1
        reboot -f
    fi
fi
