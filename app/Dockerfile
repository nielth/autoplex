FROM alpine:latest

WORKDIR /code

ADD . /code

RUN apk add cmd:pip3 git python3-dev && pip3 install --upgrade pip

RUN apk add --no-cache g++ gcc libxslt-dev

COPY requirements.txt /code/requirements.txt

RUN pip3 install -r /code/requirements.txt

CMD [ "python3", "/code/app.py" ] 