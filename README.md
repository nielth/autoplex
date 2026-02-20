## Environment setup

Copy the env template you want to use:

```sh
cp .env.production.example .env.production
cp .env.development.example .env.development
```

Used backend vars:

```text
QBT_URL
QBT_USER
QBT_PASS
OAUTH_FORWARD_URL
PLEX_URL
PLEX_TOKEN
DOMAIN
PATH_DISK
```

Used frontend/nginx var (production compose):

```text
NGINX_HOST
```

## Production

```sh
docker compose -f compose.yml up --build
```

`compose.yml` reads `.env.production` for backend and frontend containers.

## Development

Start backend + support services with development env:

```sh
docker compose -f compose.dev.yml up --build
```

`compose.dev.yml` is a standalone development stack and reads `.env.development`.

Run frontend locally:

```sh
cd frontend/src
bun install
VITE_GO_BACKEND_LOCATION=http://localhost:8080 bun dev
```
