# Use official Golang image to build the Go app
FROM golang:1.22-alpine as builder

# Set the working directory
WORKDIR /app

# Copy Go modules and install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o main .

# Use a smaller image for production
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the binary and migrations from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrations /migrations

# Expose port 8080 for the Go application
EXPOSE 8080

# Command to run the Go app
CMD ["./main"]
