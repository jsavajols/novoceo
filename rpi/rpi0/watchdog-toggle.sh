#!/bin/sh
# watchdog-toggle.sh - Suspend ou reactive le watchdog reseau (crond busybox)
# Usage: watchdog-toggle.sh suspend | resume | status
# Necessite root

CRONTAB=/etc/crontabs/root
PATTERN="/opt/novoceo/watchdog-net.sh"

case "$1" in
  suspend)
    if grep -q "^#.*${PATTERN}" "${CRONTAB}"; then
      echo "Watchdog deja suspendu"
      exit 0
    fi
    if ! grep -q "${PATTERN}" "${CRONTAB}"; then
      echo "Erreur: entree watchdog introuvable dans ${CRONTAB}"
      exit 1
    fi
    sed -i "s|^\(.* ${PATTERN}\)|#\1|" "${CRONTAB}"
    rc-service crond restart
    echo "Watchdog suspendu"
    ;;
  resume)
    if ! grep -q "^#.*${PATTERN}" "${CRONTAB}"; then
      if grep -q "^[^#].*${PATTERN}" "${CRONTAB}"; then
        echo "Watchdog deja actif"
      else
        echo "Erreur: entree watchdog introuvable dans ${CRONTAB}"
        exit 1
      fi
      exit 0
    fi
    sed -i "s|^#\(.*${PATTERN}\)|\1|" "${CRONTAB}"
    rc-service crond restart
    echo "Watchdog reactive"
    ;;
  status)
    if grep -q "^[^#].*${PATTERN}" "${CRONTAB}"; then
      echo "Watchdog: actif"
    else
      echo "Watchdog: suspendu"
    fi
    rc-service crond status
    ;;
  *)
    echo "Usage: $0 suspend | resume | status"
    exit 1
    ;;
esac
