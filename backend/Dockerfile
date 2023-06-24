FROM python:3-alpine

WORKDIR /api

COPY ./api .

RUN pip3 install -r requirements.txt

CMD [ "python3", "-u", "app.py" ]