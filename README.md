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

## Disk balancer

New downloads land on the spare disk (`DISK_BALANCER_SPARE_DISK`, default `sde`). When enabled, a worker runs every 10 minutes. It moves high bitrate torrents off the spare disk and spreads them across the pool disks (`DISK_BALANCER_POOL_DISKS`), so the pool disks stay full and the spare disk keeps as much free space as possible.

- Every move is a qBittorrent `setLocation` (`/downloads/<disk>/<category>`), so torrents keep seeding.
- A torrent counts as heavy when its biggest video file averages `DISK_BALANCER_HEAVY_MBPS` or more (default 40). The bitrate is read with `ffprobe` (installed in the backend image). If probing fails, it is guessed from the file size, assuming a movie runs 2 hours and an episode 50 minutes.
- The newest heavy torrent goes to the pool disk with the fewest heavy torrents added in the last `DISK_BALANCER_HOT_DAYS` days. If that disk is full, its oldest light torrents are moved to the spare disk first.
- When a pool disk has room and no heavy torrent needs placing, the newest light torrents that fit are moved onto it.
- Every disk keeps `DISK_BALANCER_MIN_FREE_GB` free (default 20).
- Nothing moves while Plex is streaming or while qBittorrent is still moving something. Plex libraries are rescanned after moves finish.

The backend reads free space from `/<disk>`, so every disk has to be mounted read-only (see `compose.yaml`). Start with `DISK_BALANCER_DRY_RUN=true` and check the `disk balancer:` lines in the backend logs before letting it move anything.

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
docker compose --env-file .env.production -f compose.yaml up --build
```

`compose.yml` reads `.env.production` for backend and frontend containers.

## Development

Start backend + support services with development env:

```sh
docker compose -f local.yaml up --build
```

`compose.dev.yml` is a standalone development stack and reads `.env.development`.

Run frontend locally:

```sh
cd frontend/src
bun install
VITE_GO_BACKEND_LOCATION=http://localhost:8080 bun dev
```
