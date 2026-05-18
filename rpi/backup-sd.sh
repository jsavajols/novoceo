#!/usr/bin/env bash
# backup-sd.sh - Sauvegarde et restauration de la carte SD du RPi
# Usage: brancher la carte SD sur le laptop, puis lancer ce script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKUP_DIR="${BACKUP_DIR:-${SCRIPT_DIR}/backups}"
MAX_BACKUPS=3
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

usage() {
    echo "Usage: $0 <commande> [device]"
    echo ""
    echo "Commandes:"
    echo "  detect             Détecte les périphériques SD/USB candidats"
    echo "  backup [device]    Sauvegarde la carte SD"
    echo "  restore [device]   Restaure une sauvegarde sur la carte SD"
    echo "  list               Liste les sauvegardes disponibles"
    echo ""
    echo "Variables d'environnement:"
    echo "  BACKUP_DIR         Répertoire de sauvegarde (défaut: $SCRIPT_DIR/backups)"
    echo ""
    echo "Exemples:"
    echo "  sudo $0 detect"
    echo "  sudo $0 backup /dev/sdb"
    echo "  sudo $0 restore /dev/sdb"
    echo "  sudo $0 list"
    exit 1
}

check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo "Erreur: ce script doit être exécuté en root"
        echo "  sudo $0 $*"
        exit 1
    fi
}

check_deps() {
    for cmd in dd gzip gunzip lsblk; do
        if ! command -v "$cmd" &>/dev/null; then
            echo "Erreur: commande manquante: $cmd"
            exit 1
        fi
    done
}

has_pv() {
    command -v pv &>/dev/null
}

# Liste les périphériques bloc amovibles ou de petite taille (candidats SD)
detect_devices() {
    echo "Périphériques détectés:"
    echo ""
    lsblk -d -o NAME,SIZE,TYPE,TRAN,MODEL,RM | grep -E "disk" | while read -r name size type tran model rm; do
        local dev="/dev/$name"
        local flags=""
        [[ "$rm" == "1" ]] && flags="${flags}[amovible] "
        [[ "$tran" == "usb" ]] && flags="${flags}[USB] "
        [[ "$tran" == "mmc" ]] && flags="${flags}[SD/MMC] "
        echo "  $dev  $size  $flags$model"
    done
    echo ""
    echo "Partitions montées:"
    lsblk -o NAME,SIZE,MOUNTPOINT | grep -v "^$" | grep "/" || echo "  (aucune)"
}

# Vérifie si des partitions du device sont montées et les démonte
unmount_device() {
    local device="$1"
    local mounted
    mounted=$(lsblk -ln -o MOUNTPOINT "$device" 2>/dev/null | grep -v "^$" || true)

    if [[ -n "$mounted" ]]; then
        echo "Partitions montées sur $device:"
        lsblk -o NAME,MOUNTPOINT "$device" | grep -v "^$"
        echo ""
        read -rp "Démonter automatiquement ? [oui/non]: " confirm
        [[ "$confirm" == "oui" ]] || { echo "Annulé. Démontez manuellement avant de continuer."; exit 1; }

        # Démonte toutes les partitions du device
        while IFS= read -r mountpoint; do
            [[ -z "$mountpoint" ]] && continue
            echo "  Démontage: $mountpoint"
            umount "$mountpoint"
        done <<< "$mounted"
        echo "Démontage terminé."
        echo ""
    fi
}

list_backups() {
    echo "Sauvegardes dans: $BACKUP_DIR"
    echo ""

    shopt -s nullglob
    local files=("${BACKUP_DIR}"/rpi_sd_*.img.gz)
    shopt -u nullglob

    if [[ ${#files[@]} -eq 0 ]]; then
        echo "  Aucune sauvegarde trouvée."
        return 1
    fi

    local i=1
    for f in "${files[@]}"; do
        local size
        size=$(du -h "$f" | cut -f1)
        local raw_ts
        raw_ts=$(basename "$f" .img.gz | sed 's/rpi_sd_//')
        local date_fmt
        date_fmt=$(echo "$raw_ts" | sed 's/\([0-9]\{4\}\)\([0-9]\{2\}\)\([0-9]\{2\}\)_\([0-9]\{2\}\)\([0-9]\{2\}\)\([0-9]\{2\}\)/\1-\2-\3 \4:\5:\6/')
        echo "  [$i] $date_fmt  $size  $(basename "$f")"
        ((i++))
    done
    return 0
}

rotate_backups() {
    shopt -s nullglob
    local files=("${BACKUP_DIR}"/rpi_sd_*.img.gz)
    shopt -u nullglob

    local count=${#files[@]}
    if [[ $count -gt $MAX_BACKUPS ]]; then
        local to_delete=$(( count - MAX_BACKUPS ))
        echo "Rotation: suppression de $to_delete ancienne(s) sauvegarde(s)..."
        for f in "${files[@]:0:$to_delete}"; do
            echo "  Suppression: $(basename "$f")"
            rm -f "$f"
        done
    fi
}

# Demande le device si non fourni et propose la détection
pick_device() {
    local device="${1:-}"
    if [[ -z "$device" ]]; then
        detect_devices
        read -rp "Périphérique à utiliser (ex: /dev/sdb, /dev/mmcblk0): " device
    fi
    echo "$device"
}

do_backup() {
    check_root
    local device
    device=$(pick_device "${1:-}")

    if [[ ! -b "$device" ]]; then
        echo "Erreur: $device n'est pas un périphérique bloc valide"
        exit 1
    fi

    unmount_device "$device"

    mkdir -p "$BACKUP_DIR"

    local backup_file="${BACKUP_DIR}/rpi_sd_${TIMESTAMP}.img.gz"
    local size_bytes
    size_bytes=$(blockdev --getsize64 "$device")
    local size_human
    size_human=$(numfmt --to=iec "$size_bytes" 2>/dev/null || echo "${size_bytes} octets")

    echo "Périphérique : $device ($size_human)"
    echo "Destination  : $backup_file"
    echo ""
    read -rp "Démarrer la sauvegarde ? [oui/non]: " confirm
    [[ "$confirm" == "oui" ]] || { echo "Annulé."; exit 0; }

    echo ""
    echo "Sauvegarde en cours (peut prendre plusieurs minutes)..."
    if has_pv; then
        dd if="$device" bs=4M | pv -s "$size_bytes" -i 1 -p -t -e -r -b | gzip -c > "$backup_file"
    else
        echo "(installez 'pv' pour une barre de progression)"
        dd if="$device" bs=4M status=progress | gzip -c > "$backup_file"
    fi
    sync

    echo ""
    echo "Sauvegarde terminée: $(du -h "$backup_file" | cut -f1)"

    rotate_backups
    echo ""
    list_backups
}

do_restore() {
    check_root

    echo ""
    list_backups || exit 1
    echo ""

    shopt -s nullglob
    local files=("${BACKUP_DIR}"/rpi_sd_*.img.gz)
    shopt -u nullglob
    local count=${#files[@]}

    read -rp "Numéro de la sauvegarde à restaurer [1-$count]: " choice
    if ! [[ "$choice" =~ ^[0-9]+$ ]] || (( choice < 1 || choice > count )); then
        echo "Choix invalide."
        exit 1
    fi

    local selected="${files[$((choice - 1))]}"

    local device
    device=$(pick_device "${1:-}")

    if [[ ! -b "$device" ]]; then
        echo "Erreur: $device n'est pas un périphérique bloc valide"
        exit 1
    fi

    unmount_device "$device"

    echo ""
    echo "Sauvegarde : $(basename "$selected")"
    echo "Destination: $device"
    echo ""
    echo "ATTENTION: TOUTES LES DONNEES SUR $device SERONT EFFACEES !"
    read -rp "Confirmer la restauration ? [oui/non]: " confirm
    [[ "$confirm" == "oui" ]] || { echo "Annulé."; exit 0; }

    echo ""
    echo "Restauration en cours..."
    local file_size
    file_size=$(stat -c%s "$selected")
    if has_pv; then
        pv -s "$file_size" -i 1 -p -t -e -r -b "$selected" | gunzip -c | dd of="$device" bs=4M
    else
        echo "(installez 'pv' pour une barre de progression)"
        gunzip -c "$selected" | dd of="$device" bs=4M status=progress
    fi
    sync

    echo ""
    echo "Restauration terminée."
}

check_deps

CMD="${1:-}"
case "$CMD" in
    detect)  check_root; detect_devices ;;
    backup)  do_backup "${2:-}" ;;
    restore) do_restore "${2:-}" ;;
    list)    list_backups ;;
    *)       usage ;;
esac
