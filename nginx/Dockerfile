FROM nginx:alpine
RUN rm /etc/nginx/conf.d/default.conf
COPY ./conf.d/nginx.conf /etc/nginx/conf.d
COPY ./letsencrypt/archive/nielth.com/ /etc/nginx
EXPOSE 1337
