FROM python:3-alpine

WORKDIR /api

RUN apk add --no-cache tzdata
ENV TZ=Europe/Oslo

ENV MYSQL_HOST=$MYSQL_HOST
ENV MYSQL_USER=$MYSQL_USER
ENV MYSQL_PASSWORD=$MYSQL_PASSWORD
ENV MYSQL_DATABASE=$MYSQL_DATABASE

COPY ./api .

RUN pip install -r requirements.txt

CMD [ "python", "-u", "api.py" ]