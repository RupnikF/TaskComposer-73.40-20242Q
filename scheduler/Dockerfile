FROM golang:1.23 as builder

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o ./app .

FROM alpine:3.20

WORKDIR /usr/src/app

COPY --from=builder ./app .
COPY ./*.sh ./

RUN chmod +x ./*.sh

CMD ["./start.sh"]
