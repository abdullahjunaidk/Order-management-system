# Use the official Go image as the base image
FROM golang:1.25.5-bookworm

# Install essential build tools
RUN apt-get update && apt-get install -y make curl

# Set the working directory
WORKDIR /app

# Copy the entire project
COPY . .

# Install development tools
RUN go install github.com/air-verse/air@latest \
    && go install github.com/swaggo/swag/cmd/swag@latest

# Install common service dependencies
WORKDIR /app/common
RUN go mod download

# Install gateway service dependencies
WORKDIR /app/gateway
RUN go mod download

# Start the application with air for hot-reloading
CMD ["air", "-c", ".air.toml"]
