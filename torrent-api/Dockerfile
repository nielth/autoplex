FROM python:3-alpine

RUN apk add git && \
    git clone https://github.com/Ryuk-me/Torrent-Api-py

WORKDIR /Torrent-Api-py

RUN pip3 install -r requirements.txt

EXPOSE 8080

CMD [ "python", "-u", "main.py" ]