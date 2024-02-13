## Prerequisites

`backend/api` requires `cookies.json` from torrentleech.

## For Prod (no trailing slash)

`.env`
```
QBT_USER=<qbt_username>
QBT_PASS=<qbt_pass>
JWT_SECRET=<uuid4_random_secret>
OAUTH_FORWARD_URL=https://autoplex.example.com
PLEX_URL=http://plex.example.com:32400
```

```
docker compose -f docker-compose.yml up --build
```

## For Dev

```
docker compose -f docker-compose-dev.yml up --build
```

