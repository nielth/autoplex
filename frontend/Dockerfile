FROM oven/bun:1 AS build
WORKDIR /app
COPY src/package.json src/bun.lockb ./
RUN bun install --frozen-lockfile
COPY src/ .
RUN bun run build

FROM caddy:2-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=build /app/dist /srv
