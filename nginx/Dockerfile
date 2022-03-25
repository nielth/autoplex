FROM nginx:alpine
RUN rm /etc/nginx/conf.d/default.conf && \
    apk update && \
    apk add --update certbot-nginx && \
    mkdir /etc/letsencrypt && \
    crontab -l | { cat; echo "0 12 * * * /usr/bin/certbot renew --quiet"; } | crontab -
COPY ./conf.d/nginx.conf /etc/nginx/conf.d
COPY ./letsencrypt/archive/nielth.com/ /etc/nginx
EXPOSE 1337
    
#certbot --nginx -d yourdomain.com -d www.yourdomain.com && \
