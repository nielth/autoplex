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
PLEX_TOKEN=<token>
NGINX_HOST=nginx.example.com
QBT_URL=https://qbt.example.com
```

```
docker compose -f docker-compose.yml up --build
```

## For Dev

```sh
docker compose -f docker-compose-dev.yml up --build
```

For running react directly on machine
```sh
docker compose -f docker-compose-dev.yml up flask-api --build
yarn && REACT_APP_FLASK_LOCATION=http://localhost:5050 yarn start
```

`.env`
```
QBT_USER=admin
QBT_PASS=adminadmin
JWT_SECRET=developing_secret
OAUTH_FORWARD_URL=http://localhost:3000
REACT_APP_FLASK_LOCATION=http://localhost:5050
PLEX_URL=http://nielth.com:32444
PLEX_TOKEN=<token>
```

