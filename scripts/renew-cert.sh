#!/usr/bin/env bash
# Issue or renew the TLS cert for NGINX_HOST via certbot webroot, then copy the
# result into frontend/nginx/ssl/ (where nginx reads from) and reload node-prod.
#
# Usage:
#   ./scripts/renew-cert.sh                  # normal renewal, real CA
#   ./scripts/renew-cert.sh --force          # force renewal even if cert is fresh
#   ./scripts/renew-cert.sh --staging        # staging CA (no rate limits, fake cert)
#   flags combine: ./scripts/renew-cert.sh --force --staging
#
# Requirements:
#   - .env.production contains NGINX_HOST and CERTBOT_EMAIL
#   - node-prod is running so nginx serves /.well-known/acme-challenge/
#   - port 80 reachable from the internet at NGINX_HOST
#
# Safe to re-run; non-interactive.

set -euo pipefail

cd "$(dirname "$0")/.."

ENV_FILE=".env.production"
if [[ ! -f "$ENV_FILE" ]]; then
    echo "error: $ENV_FILE not found" >&2
    exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

: "${NGINX_HOST:?set NGINX_HOST in $ENV_FILE}"
: "${CERTBOT_EMAIL:?set CERTBOT_EMAIL in $ENV_FILE}"

FORCE_FLAG=""
STAGING_FLAG=""
for arg in "$@"; do
    case "$arg" in
        --force) FORCE_FLAG="--force-renewal" ;;
        --staging) STAGING_FLAG="--staging" ;;
        *) echo "unknown flag: $arg" >&2; exit 1 ;;
    esac
done

REPO_ROOT="$(pwd)"
LIVE_DIR="frontend/certbot/conf/live/$NGINX_HOST"
TARGET_DIR="frontend/nginx/ssl"

if ! docker compose --env-file "$ENV_FILE" ps --services --filter "status=running" | grep -q '^node-prod$'; then
    echo "error: node-prod is not running; nginx must be up to serve the ACME challenge." >&2
    echo "       run: docker compose --env-file $ENV_FILE up -d" >&2
    exit 1
fi

echo "renew-cert: host=$NGINX_HOST email=$CERTBOT_EMAIL force=${FORCE_FLAG:-no} staging=${STAGING_FLAG:-no}"
echo "renew-cert: running certbot via webroot"
docker run --rm --name certbot-renew \
    -v "$REPO_ROOT/frontend/certbot/conf:/etc/letsencrypt" \
    -v "$REPO_ROOT/frontend/certbot/www:/var/www/certbot" \
    certbot/certbot certonly --webroot -w /var/www/certbot \
    -d "$NGINX_HOST" \
    --email "$CERTBOT_EMAIL" \
    --agree-tos --no-eff-email --non-interactive \
    $FORCE_FLAG $STAGING_FLAG

if [[ ! -f "$LIVE_DIR/fullchain.pem" ]]; then
    echo "error: expected cert not found at $LIVE_DIR/fullchain.pem" >&2
    exit 1
fi

echo "renew-cert: copying cert into $TARGET_DIR/"
sudo cp -L "$LIVE_DIR/fullchain.pem" "$TARGET_DIR/fullchain.pem"
sudo cp -L "$LIVE_DIR/privkey.pem"  "$TARGET_DIR/privkey.pem"
sudo chown "$USER":"$USER" "$TARGET_DIR/fullchain.pem" "$TARGET_DIR/privkey.pem"

echo "renew-cert: reloading nginx"
docker compose --env-file "$ENV_FILE" exec node-prod nginx -s reload

echo "renew-cert: done. new expiry:"
openssl x509 -in "$TARGET_DIR/fullchain.pem" -noout -subject -dates
