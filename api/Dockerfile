FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy files from the build context.
# Pants will place the files preserving their path structure relative to the build root.
# So we expect to see api/go.mod, api/main.go, api/internal/..., etc.
COPY . .

# Change to the directory containing main.go and go.mod
WORKDIR /app/api

RUN go mod download

# Build the binary
# We build for the architecture of the builder container (which matches the --platform flag in docker buildx)
RUN go build -o /app/aquascore main.go

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/aquascore .
RUN apk add --no-cache ca-certificates

# Expose port if needed (e.g. 8080)
EXPOSE 8080

CMD ["./aquascore", "server"]
