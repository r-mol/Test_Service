# Stage 1: Build the Go application
FROM golang:1.22.5 as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Download Go modules
RUN go mod download

# Build the Go application as a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app ./cmd/main.go

# Stage 2: Create a minimal image to run the Go application
FROM alpine:latest

RUN apk --no-cache add ca-certificates

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the binary file from the previous stage
COPY --from=builder /app/app .

# Ensure the app binary is executable
RUN chmod +x app

# Expose port 8081 to the outside world
EXPOSE 8081

# Command to run the executable
CMD ["./app", "start", "--config", "/app/config.yaml", "--jwt_key", "$JWT_KEY"]