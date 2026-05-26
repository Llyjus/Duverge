FROM golang:1.26-alpine

WORKDIR /app

COPY . .

RUN go mod download

RUN go build \
-o game \
./cmd/main.go

CMD ["./game"]


