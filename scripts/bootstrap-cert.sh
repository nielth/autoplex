#!/usr/bin/env bash
# One-off TLS cert issuance via certbot standalone mode.
#
# Use this the first time you bring up prod, or any time the existing cert
# in frontend/certbot/conf has been lost. Routine renewals are handled by
# the `certbot` service in compose.yml.
#
# Usage:
#   NGINX_HOST=autoplex.example.com CERTBOT_EMAIL=you@example.com ./scripts/bootstrap-cert.sh            # staging
#   NGINX_HOST=autoplex.example.com CERTBOT_EMAIL=you@example.com ./scripts/bootstrap-cert.sh --prod     # real cert
#
# NGINX_HOST and CERTBOT_EMAIL can also live in .env.production.
#
# Requirements: port 80 reachable from the public internet, DNS pointing here,
# nothing else bound to host port 80 while the script runs.

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

: "${NGINX_HOST:?set NGINX_HOST in $ENV_FILE or env}"
: "${CERTBOT_EMAIL:?set CERTBOT_EMAIL in $ENV_FILE or env}"

STAGING_FLAG="--staging"
MODE_LABEL="STAGING (test cert, no rate limits)"
if [[ "${1:-}" == "--prod" ]]; then
    STAGING_FLAG=""
    MODE_LABEL="PROD (real Let's Encrypt cert)"
fi

echo "bootstrap-cert: host=$NGINX_HOST email=$CERTBOT_EMAIL mode=$MODE_LABEL"
echo "bootstrap-cert: stopping compose stack so port 80 is free"
docker compose --env-file "$ENV_FILE" down

echo "bootstrap-cert: running certbot standalone"
docker compose --env-file "$ENV_FILE" run --rm \
    -p 80:80 \
    --entrypoint certbot \
    certbot certonly --standalone \
    -d "$NGINX_HOST" \
    --email "$CERTBOT_EMAIL" \
    --agree-tos --no-eff-email --non-interactive \
    $STAGING_FLAG

CERT_DIR="frontend/certbot/conf/live/$NGINX_HOST"
if [[ ! -f "$CERT_DIR/fullchain.pem" ]]; then
    echo "error: expected cert not found at $CERT_DIR/fullchain.pem" >&2
    exit 1
fi

echo "bootstrap-cert: cert written to $CERT_DIR"
if [[ -n "$STAGING_FLAG" ]]; then
    echo "bootstrap-cert: this was a STAGING cert (browsers will warn)."
    echo "bootstrap-cert: when ready, re-run with --prod for the real cert."
else
    echo "bootstrap-cert: bring the stack back up with:"
    echo "  docker compose --env-file $ENV_FILE up -d"
fi
