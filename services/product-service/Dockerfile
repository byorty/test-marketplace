FROM golang:1.25 AS builder

WORKDIR /app

COPY services/ .

WORKDIR /app/product-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o product-service \
    ./cmd/product-service


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/product-service .
COPY --from=builder /app/product-service/config.yaml ./config/config.yaml
COPY --from=builder /app/common/rbac ./common/rbac
COPY --from=builder /app/keys ./keys

EXPOSE 8080

CMD ["./product-service"]