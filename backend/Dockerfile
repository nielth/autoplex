FROM python:3-alpine

WORKDIR /api

RUN apk add --no-cache tzdata
# Set unbuffered output for print statements
ENV PYTHONUNBUFFERED=1
ENV TZ=Europe/Oslo

COPY ./api .

RUN pip install -r requirements.txt
CMD ["gunicorn", "-b", "0.0.0.0:5000", "api:app"]
