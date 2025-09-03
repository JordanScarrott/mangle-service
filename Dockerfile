# Stage 1: Build the Go binary
FROM golang:1.18-alpine AS builder

WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mangle-microservice ./cmd/mangle-microservice

# Stage 2: Create the final, minimal image
FROM alpine:latest

# Copy the built binary from the builder stage
COPY --from=builder /app/mangle-microservice .

# Set the command to run the application
CMD ["./mangle-microservice"]
