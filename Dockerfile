FROM alpine:latest

ADD . /code

WORKDIR /code

RUN apk add cmd:pip3 git python3-dev && pip3 install --upgrade pip

RUN apk add --no-cache g++ gcc libxslt-dev

COPY requirements.txt /code/requirements.txt

EXPOSE 5000

RUN pip3 install -r /code/requirements.txt

CMD [ "python3", "app/app.py" ] 