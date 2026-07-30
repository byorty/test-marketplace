FROM golang:1.25 AS builder

WORKDIR /app

COPY services/order-service .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o order-service \
    ./cmd/order-service

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/order-service .
COPY --from=builder /app/config.yaml .

EXPOSE 8081

CMD ["./order-service"]