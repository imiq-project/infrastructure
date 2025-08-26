FROM golang:1.25 AS builder
WORKDIR /app/src
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY ./src .
RUN go build -o collector

FROM debian:13
WORKDIR /app/src
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
    && apt-get clean
RUN update-ca-certificates
COPY --from=builder /app/src/collector .
COPY ./config.yml /app
CMD ["./collector", "--config", "/app/config.yml"]
