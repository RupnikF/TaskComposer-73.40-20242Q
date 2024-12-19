FROM python:3.12

WORKDIR /app

COPY requirements.txt .

RUN pip install -r requirements.txt

COPY service.py .

ENTRYPOINT [ "python3" ]

CMD [ "service.py" ]