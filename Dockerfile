# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

# Install tzdata and ca-certificates for security and timezone handling
RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy dependency definition files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the Go application binary for Linux container environment
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Stage 2: Final minimal image
FROM alpine:3.19

# Install security certificates and timezone database
RUN apk update && apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/entrypoint.sh .

# Copy environment example to serve as a fallback or template
COPY --from=builder /app/.env.example .env

RUN sed -i 's/\r$//' entrypoint.sh && chmod +x entrypoint.sh

# Expose the API port
EXPOSE 4001

# Execute the application via entrypoint
ENTRYPOINT ["./entrypoint.sh"]

