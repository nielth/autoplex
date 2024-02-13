FROM node:18-alpine AS builder
WORKDIR /app
COPY . .

RUN yarn  
RUN yarn build

FROM nginx:alpine
COPY --from=builder /app/build /var/www/myapp
COPY ./nginx/ssl /etc/nginx/ssl
COPY ./nginx/conf /etc/nginx/templates
# ENTRYPOINT ["tail", "-f", "/dev/null"]
