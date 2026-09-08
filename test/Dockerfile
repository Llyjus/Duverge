FROM python:3.13-slim

WORKDIR /test

COPY requirements.txt .

RUN pip install \
--no-cache-dir \
-r requirements.txt

COPY . .

CMD ["pytest","-s"]