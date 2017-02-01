FROM python:3

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app
COPY . /usr/src/app

RUN pip install --no-cache-dir -r requirements.txt
EXPOSE 8000

CMD["python","route.py"]