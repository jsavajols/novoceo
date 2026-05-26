#!/bin/sh
# watchdog-toggle.sh - Suspend ou reactive le watchdog reseau (systemd timer)
# Usage: watchdog-toggle.sh suspend | resume | status
# Necessite sudo/root

case "$1" in
  suspend)
    if ! systemctl is-active --quiet watchdog-net.timer; then
      echo "Watchdog deja suspendu"
      exit 0
    fi
    systemctl stop watchdog-net.timer
    echo "Watchdog suspendu"
    ;;
  resume)
    if systemctl is-active --quiet watchdog-net.timer; then
      echo "Watchdog deja actif"
      exit 0
    fi
    systemctl start watchdog-net.timer
    echo "Watchdog reactive"
    ;;
  status)
    if systemctl is-active --quiet watchdog-net.timer; then
      echo "Watchdog: actif"
    else
      echo "Watchdog: suspendu"
    fi
    systemctl status watchdog-net.timer --no-pager -l
    ;;
  *)
    echo "Usage: $0 suspend | resume | status"
    exit 1
    ;;
esac
