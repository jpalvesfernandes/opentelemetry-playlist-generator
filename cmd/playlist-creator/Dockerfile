FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
WORKDIR /app/cmd/playlist-creator
RUN go build -o /playlist-creator

# Create a lightweight image
FROM alpine:latest
COPY --from=builder /playlist-creator /playlist-creator
EXPOSE 8082
ENTRYPOINT ["/playlist-creator"]
