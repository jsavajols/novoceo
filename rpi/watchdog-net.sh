#!/usr/bin/env bash
# watchdog-net.sh - Reboot si la connectivité réseau est dégradée
# Cron : * * * * * root /opt/novoceo/watchdog-net.sh

TARGET="8.8.8.8"
COUNT=10
THRESHOLD=50  # % d'erreurs au-delà duquel on reboot
LOG="/var/log/watchdog-net.log"

lost=$(ping -c "$COUNT" -q "$TARGET" 2>/dev/null \
    | awk '/packets transmitted/ { gsub(/%/,"",$6); print $6 }')

if [[ -z "$lost" ]]; then
    echo "$(date '+%F %T') ERREUR: ping injoignable ou sortie inattendue" >> "$LOG"
    lost=100
fi

echo "$(date '+%F %T') perte=${lost}%" >> "$LOG"

if (( lost > THRESHOLD )); then
    echo "$(date '+%F %T') REBOOT (perte=${lost}% > ${THRESHOLD}%)" >> "$LOG"
    reboot -f
fi
