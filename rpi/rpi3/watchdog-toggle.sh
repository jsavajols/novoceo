#!/bin/sh
# watchdog-toggle.sh - Suspend ou reactive le watchdog reseau (systemd timer)
# La suspension persiste apres reboot via disable/enable du timer.
# Usage: watchdog-toggle.sh suspend | resume | status
# Necessite sudo/root

is_enabled() {
    systemctl is-enabled --quiet watchdog-net.timer 2>/dev/null
}

case "$1" in
  suspend)
    if ! is_enabled; then
      echo "Watchdog deja suspendu"
      exit 0
    fi
    systemctl stop watchdog-net.timer
    systemctl disable watchdog-net.timer
    echo "Watchdog suspendu (persistera apres reboot)"
    ;;
  resume)
    if is_enabled; then
      echo "Watchdog deja actif"
      exit 0
    fi
    systemctl enable --now watchdog-net.timer
    echo "Watchdog reactive"
    ;;
  status)
    if is_enabled; then
      echo "Watchdog: actif"
    else
      echo "Watchdog: suspendu (persistera apres reboot)"
    fi
    systemctl status watchdog-net.timer --no-pager -l
    ;;
  *)
    echo "Usage: $0 suspend | resume | status"
    exit 1
    ;;
esac
