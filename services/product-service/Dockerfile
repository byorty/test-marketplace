FROM golang:1.25 AS builder

WORKDIR /app

COPY services/product-service .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o product-service \
    ./cmd/product-service

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/product-service .
COPY --from=builder /app/config.yaml .

EXPOSE 8080

CMD ["./product-service"]