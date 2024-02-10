FROM python:3-alpine

WORKDIR /api

RUN apk add --no-cache tzdata
ENV TZ=Europe/Oslo

ENV JWT_SECRET=$JWT_SECRET
ENV QBT_USER=$QBT_USER
ENV QBT_PASS=$QBT_PASS

COPY ./api .

RUN pip install -r requirements.txt

CMD [ "python", "-u", "api.py" ]