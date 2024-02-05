# FROM node:18-alpine AS builder
# WORKDIR /app
# COPY ./frontend .

# RUN yarn  
# RUN yarn build

FROM nginx:alpine
COPY ./nginx/ /etc/nginx/conf.d/
# COPY --from=builder /app/build /var/www/myapp
# ENTRYPOINT ["tail", "-f", "/dev/null"]
