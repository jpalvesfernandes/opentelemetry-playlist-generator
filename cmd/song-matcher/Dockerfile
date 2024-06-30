FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
WORKDIR /app/cmd/song-matcher
RUN go build -o /song-matcher

# Create a lightweight image
FROM alpine:latest
COPY --from=builder /song-matcher /song-matcher
EXPOSE 8081
ENTRYPOINT ["/song-matcher"]
