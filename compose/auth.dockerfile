# Use the official Go image as the base image
FROM golang:1.25.5-bookworm

# Install essential build tools
RUN apt-get update && apt-get install -y make curl

# Install grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.14 && \
    curl -fL -o /bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

# Set the working directory
WORKDIR /app

# Copy the entire project
COPY . .

# Install development tools
RUN go install github.com/air-verse/air@latest

# Install common service dependencies
WORKDIR /app/common
RUN go mod download

# Install auth service dependencies
WORKDIR /app/auth
RUN go mod download

# Start the application with air for hot-reloading
CMD ["air", "-c", ".air.toml"]