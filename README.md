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
MYSQL_HOST
MYSQL_PORT
MYSQL_DATABASE
MYSQL_USER
MYSQL_PASSWORD
```

Used frontend/nginx var (production compose):

```text
NGINX_HOST
```

## Shared MySQL (single DB for dev + prod)

This project now writes these MySQL audit tables automatically at startup:

```text
users
admin_users
login_events
search_events
download_events
download_delete_requests
```

Grant admin manually by inserting the internal user id into `admin_users`:

```sql
INSERT INTO admin_users (user_id) VALUES (123);
```

Set the same `MYSQL_*` values in both `.env.development` and `.env.production` if you want one shared database for all environments.

`compose.yml` now includes a MySQL container (`mysql:8.4`) and stores data on the host at `${HOME}/mysql` (your `~/mysql`).

In production compose:

- Set `MYSQL_HOST=mysql` so backend resolves the MySQL service by Docker DNS.
- `MYSQL_ROOT_PASSWORD` is required by MySQL startup.

For development on another machine:

- Point `MYSQL_HOST` in `.env.development` to the LAN IP of the machine running `compose.yml`.
- Keep `MYSQL_PORT=3306` (or match your published port).

## Production

```sh
docker compose --env-file .env.production -f compose.yml up --build
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
