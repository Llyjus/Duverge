FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build \
-o gateway \
./cmd/server/main.go

EXPOSE 26526
EXPOSE 26527

CMD ["./gateway"]