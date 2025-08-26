FROM golang:1.25 AS builder
WORKDIR /app/src
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY ./src .
RUN go build -o collector

FROM debian:13
WORKDIR /app/src
COPY --from=builder /app/src/collector .
COPY ./config.yml /app
CMD ["./collector", "--config", "/app/config.yml"]
