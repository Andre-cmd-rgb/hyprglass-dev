#!/usr/bin/env bash
set -euo pipefail
APN=""
SIM_PIN=""
CONNECTION_NAME="Hyprglass LTE"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --apn) APN=${2:-}; shift 2 ;;
    --pin) SIM_PIN=${2:-}; shift 2 ;;
    --name) CONNECTION_NAME=${2:-}; shift 2 ;;
    *) echo "unknown option: $1"; exit 2 ;;
  esac
done

ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
install -Dm755 "$ROOT/scripts/hyprglass-modem-autounlock.sh" /usr/local/lib/hyprglass/modem-autounlock.sh
install -d -m 700 /etc/hyprglass
umask 077
{
  printf 'SIM_PIN=%q\n' "$SIM_PIN"
  printf 'APN=%q\n' "$APN"
  printf 'CONNECTION_NAME=%q\n' "$CONNECTION_NAME"
} > /etc/hyprglass/modem.env
chmod 600 /etc/hyprglass/modem.env
cat > /etc/systemd/system/hyprglass-modem-autounlock.service <<'UNIT'
[Unit]
Description=Hyprglass modem SIM unlock and autoconnect
After=NetworkManager.service ModemManager.service
Wants=NetworkManager.service ModemManager.service

[Service]
Type=oneshot
ExecStart=/usr/local/lib/hyprglass/modem-autounlock.sh

[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload
systemctl enable --now hyprglass-modem-autounlock.service
