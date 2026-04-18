FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /payment-service ./payment-service

FROM alpine:3.20
COPY --from=builder /payment-service /payment-service
EXPOSE 8081
CMD ["/payment-service"]