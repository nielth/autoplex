FROM alpine:latest
WORKDIR /code
COPY . /code
RUN apk add --update cmd:pip3 git python3-dev && \
    apk add --no-cache g++ gcc libxslt-dev && \
    pip3 install --upgrade pip && \
    pip3 install -r /code/requirements.txt
CMD [ "python3", "/code/app.py" ] 