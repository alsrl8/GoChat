# Build the Go binary
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copy source
COPY . .

# Build (static where possible)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./

# Runtime image
FROM alpine:3.20
WORKDIR /app

# Copy binary
COPY --from=builder /app/server /app/server

# Run the application
ENTRYPOINT ["/app/server"]
