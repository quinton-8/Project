# Build stage
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY go.mod ./
# COPY go.sum ./ # Uncomment when you add external dependencies
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /taifacare ./cmd/server

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /taifacare /taifacare
EXPOSE 8080
ENTRYPOINT ["/taifacare"]