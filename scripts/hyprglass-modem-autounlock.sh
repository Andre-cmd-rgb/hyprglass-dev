#!/usr/bin/env bash
set -euo pipefail

ENV_FILE=${HYPRGLASS_MODEM_ENV:-/etc/hyprglass/modem.env}
if [[ -r "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  source "$ENV_FILE"
fi

SIM_PIN=${SIM_PIN:-}
APN=${APN:-}
CONNECTION_NAME=${CONNECTION_NAME:-Hyprglass LTE}

log(){ printf 'hyprglass-modem: %s\n' "$*"; }
need(){ command -v "$1" >/dev/null 2>&1 || { log "$1 missing"; exit 0; }; }

need mmcli
need nmcli

nmcli radio wwan on >/dev/null 2>&1 || true
modems=$(mmcli -L 2>/dev/null | sed -n 's#.*Modem/\([0-9][0-9]*\).*#\1#p' | sort -u)
[[ -n "$modems" ]] || { log "no modem detected"; exit 0; }

for modem in $modems; do
  info=$(mmcli -m "$modem" 2>/dev/null || true)
  sim=$(printf '%s\n' "$info" | sed -n "s#.*SIM/[[:space:]]*\([0-9][0-9]*\).*#\1#p" | head -1)
  lock=$(printf '%s\n' "$info" | sed -n "s/.*lock:[[:space:]]*'\([^']*\)'.*/\1/p" | head -1)

  if [[ "$lock" == "sim-pin" && -n "$SIM_PIN" && -n "$sim" ]]; then
    log "unlocking SIM $sim for modem $modem"
    mmcli -i "$sim" --pin="$SIM_PIN" >/dev/null || true
  fi

  if [[ -n "$APN" ]]; then
    if nmcli -t -f NAME,TYPE connection show | awk -F: -v n="$CONNECTION_NAME" '$1==n && $2=="gsm" {found=1} END {exit !found}'; then
      nmcli connection modify "$CONNECTION_NAME" \
        gsm.apn "$APN" \
        connection.autoconnect yes \
        connection.autoconnect-priority 100 \
        connection.autoconnect-retries -1 \
        ipv6.method ignore >/dev/null
      [[ -n "$SIM_PIN" ]] && nmcli connection modify "$CONNECTION_NAME" gsm.pin "$SIM_PIN" gsm.pin-flags 0 >/dev/null || true
    else
      args=(connection add type gsm ifname "*" con-name "$CONNECTION_NAME" apn "$APN")
      [[ -n "$SIM_PIN" ]] && args+=(pin "$SIM_PIN")
      nmcli "${args[@]}" >/dev/null || true
      nmcli connection modify "$CONNECTION_NAME" \
        connection.autoconnect yes \
        connection.autoconnect-priority 100 \
        connection.autoconnect-retries -1 \
        ipv6.method ignore >/dev/null || true
    fi
  fi
done

log "done"
