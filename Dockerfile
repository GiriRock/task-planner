# syntax=docker/dockerfile:1

FROM golang:1.23.2

# Set working directory inside the container
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project to the container, including the cmd directory
COPY . .

# Build the Go application
RUN go build -o main ./cmd/main.go

# Expose the port your app will listen on
EXPOSE 42069

# Run the binary
CMD ["./main"]

