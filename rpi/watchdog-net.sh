#!/usr/bin/env bash
# watchdog-net.sh - Reboot si la connectivité réseau est dégradée
# Géré par systemd : watchdog-net.timer + watchdog-net.service

TARGET="8.8.8.8"
COUNT=10
THRESHOLD=50  # % d'erreurs au-delà duquel on reboot

lost=$(ping -c "$COUNT" -q "$TARGET" 2>/dev/null \
    | awk '/packets transmitted/ { gsub(/%/,"",$6); print $6 }')

if [[ -z "$lost" ]]; then
    echo "ERREUR: ping injoignable ou sortie inattendue"
    lost=100
fi

echo "perte=${lost}%"

if (( lost > THRESHOLD )); then
    echo "REBOOT déclenché (perte=${lost}% > ${THRESHOLD}%)"
    reboot -f
fi
