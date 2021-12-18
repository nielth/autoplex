FROM ubuntu

ADD . /code

WORKDIR /code

RUN apt-get update -y && \
    apt-get install -y python3 python3-pip git

COPY requirements.txt /code/requirements.txt

EXPOSE 5000

RUN pip3 install -r /code/requirements.txt

CMD [ "python3", "app.py" ] 
