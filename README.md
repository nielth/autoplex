## For Prod

`.env`
```
QBT_USER=<qbt_username>
QBT_PASS=<qbt_pass>
JWT_SECRET=<uuid4 secret>
OAUTH_FORWARD_URL=<flask_api_url>/callback
```

```
docker compose -f docker-compose.yml up --build
```

## For Dev

```
docker compose -f docker-compose-dev.yml up --build
```

