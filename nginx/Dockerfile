FROM nginx:alpine
RUN rm /etc/nginx/conf.d/default.conf
COPY ./conf.d/nginx.conf /etc/nginx/conf.d
EXPOSE 1337