FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY user_customer/ ./user_customer/
COPY gateway/ ./gateway/
WORKDIR /src/gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/gateway ./cmd/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/gateway .
EXPOSE 8080
CMD ["./gateway"]
