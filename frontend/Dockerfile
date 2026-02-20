FROM oven/bun:latest AS builder
WORKDIR /app
COPY ./src .

RUN bun install 
RUN bun vite build

FROM nginx:alpine
RUN rm -f /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /var/www/myapp
COPY ./nginx/ssl /etc/nginx/ssl
COPY ./nginx/conf /etc/nginx/templates
# ENTRYPOINT ["tail", "-f", "/dev/null"]
