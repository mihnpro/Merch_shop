FROM golang:1.25-alpine AS builder
# Build context is services/ (parent of order/)
WORKDIR /build

COPY order/go.mod order/
COPY cart/go.mod cart/
COPY invetory/go.mod invetory/
COPY user/go.mod user/

WORKDIR /build/order
RUN GOFLAGS=-mod=mod go mod download

WORKDIR /build
COPY order/ order/
COPY cart/ cart/
COPY invetory/ invetory/
COPY user/ user/

WORKDIR /build/order
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./cmd/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 8086
EXPOSE 50056
CMD ["./service"]
