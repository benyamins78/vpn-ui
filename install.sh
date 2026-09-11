#!/usr/bin/env bash
# Canonical fork entrypoint. Keep deployment logic in deploy.sh so upgrades,
# backups, systemd handling, checksum verification, and BBR behavior remain one
# implementation. When fetched through process substitution there is no local
# sibling, so fetch deploy.sh from the same explicitly named fork.
set -euo pipefail
SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
if [[ -f "$SCRIPT_DIR/deploy.sh" ]]; then
    exec bash "$SCRIPT_DIR/deploy.sh" "$@"
fi
: "${VPN_UI_REPO:?set VPN_UI_REPO=owner/repository for a remote fork install}"
VPN_UI_BRANCH="${VPN_UI_BRANCH:-codex/unified-control-plane}"
VPN_UI_DEPLOY_URL="${VPN_UI_DEPLOY_URL:-https://raw.githubusercontent.com/${VPN_UI_REPO}/${VPN_UI_BRANCH}/deploy.sh}"
exec bash <(curl -fsSL "$VPN_UI_DEPLOY_URL") "$@"
