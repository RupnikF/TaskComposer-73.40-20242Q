FROM golang:1.23

WORKDIR /usr/src/app

COPY ./wait-for-it.sh .

ADD ./wait-for-it.sh  /usr/local/bin/wait-for-it
RUN chmod +x /usr/local/bin/wait-for-it

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o ./app .

RUN chmod +x ./start.sh

CMD ["./start.sh"]
