# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /dashboard ./cmd/dashboard

# Run stage — small final image, no Go toolchain included
FROM alpine:latest

WORKDIR /app
COPY --from=builder /dashboard .

EXPOSE 8080
CMD ["./dashboard"]
