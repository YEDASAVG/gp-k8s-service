FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /order-service ./order-service


FROM alpine:3.20
COPY --from=builder /order-service /order-service
EXPOSE 8080
CMD ["/order-service"]