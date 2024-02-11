FROM node:18-alpine AS builder
WORKDIR /app
COPY ./frontend .

RUN yarn  
RUN yarn build

FROM nginx:alpine
COPY ./nginx/default.conf /etc/nginx/conf.d/
COPY ./nginx/nielth.com.conf /etc/nginx/conf.d/
COPY ./nginx/ssl /etc/nginx/ssl
COPY --from=builder /app/build /var/www/myapp
# ENTRYPOINT ["tail", "-f", "/dev/null"]
