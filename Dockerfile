# Use the official Golang image to create a build artifact.
FROM golang:1.24 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server ./cmd/server

# Use a minimal alpine image for the final stage
FROM alpine:3.18

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/server /server

# Run the web service on container startup
CMD ["/server"]
