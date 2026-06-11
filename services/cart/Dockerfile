FROM golang:1.25-alpine AS builder
# Build context is services/ (parent of cart/)
WORKDIR /build

# Copy all needed modules
COPY cart/go.mod cart/go.sum cart/
COPY invetory/go.mod invetory/go.sum invetory/
COPY products/go.mod products/go.sum products/

WORKDIR /build/cart
RUN go mod download

WORKDIR /build
COPY cart/ cart/
COPY invetory/ invetory/
COPY products/ products/

WORKDIR /build/cart
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service ./cmd/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service .
EXPOSE 50055
CMD ["./service"]
