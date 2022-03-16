FROM nginx:alpine
RUN rm /etc/nginx/conf.d/default.conf && \
    apk update 
COPY ./conf.d/nginx.conf /etc/nginx/conf.d
COPY ./cert/* /etc/nginx
RUN apk add --update certbot-nginx && \
    mkdir /etc/letsencrypt
    #certbot --nginx -d yourdomain.com -d www.yourdomain.com && \
    #crontab -l | { cat; echo "0 12 * * * /usr/bin/certbot renew --quiet"; } | crontab -
EXPOSE 1337
