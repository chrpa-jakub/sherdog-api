FROM golang:latest

WORKDIR /app

COPY . .
RUN go mod download

RUN go build -o /sherdog-api ./cmd/sherdog-api

EXPOSE 8080

CMD ["/sherdog-api"]
