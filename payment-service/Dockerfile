FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY payment-service/ ./payment-service/
WORKDIR /app/payment-service
RUN go build -o payment-service .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/payment-service/payment-service .
EXPOSE 8081
CMD ["./payment-service"]
