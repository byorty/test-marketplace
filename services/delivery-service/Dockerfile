FROM golang:1.25 AS builder

WORKDIR /app

COPY services/ .

WORKDIR /app/delivery-service

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o delivery-service \
    ./cmd/delivery-service


FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/delivery-service .
COPY --from=builder /app/delivery-service/config.yaml ./config/config.yaml
COPY --from=builder /app/common/rbac ./common/rbac
COPY --from=builder /app/keys ./keys

EXPOSE 8082

CMD ["./delivery-service"]