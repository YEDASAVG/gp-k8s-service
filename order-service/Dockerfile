FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY order-service/ ./order-service/
WORKDIR /app/order-service
RUN go build -o order-service .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/order-service/order-service .
EXPOSE 8080
CMD ["./order-service"]
