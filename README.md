## Prerequisites

## For Prod (no trailing slash)

`.env`
```
QBT_USER=<qbt_username>
QBT_PASS=<qbt_pass>
QBT_URL=https://qbt.example.com
OAUTH_FORWARD_URL=https://autoplex.example.com
PLEX_URL=http://plex.example.com:32400
PLEX_TOKEN=<token>
NGINX_HOST=nginx.example.com

VITE_FLASK_LOCATION=https://example.com
```

```
docker compose -f docker-compose.yml up --build
```

## For Dev

```sh
docker compose -f docker-compose-dev.yml up --build
bun install && VITE_FLASK_LOCATION=http://localhost:5050 bun dev
```

`.env`
```
QBT_USER=admin
QBT_PASS=adminadmin
QBT_URL=https://qbt.internal.example.com
PLEX_TOKEN=<token>
OAUTH_FORWARD_URL=http://localhost:3000
PLEX_URL=http://external.example.com:32444
```

