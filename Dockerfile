FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

# Runtime image
FROM alpine:3.20
WORKDIR /app

# Copy binary
COPY --from=builder /app/server /app/server

EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/server"]
