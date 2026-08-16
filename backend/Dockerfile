FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire backend
COPY . .

# Build the main server binary
RUN go build -o server ./cmd/server/main.go
# Build the seed script binary
RUN go build -o seed ./scripts/seed_slots.go

FROM alpine:latest

WORKDIR /app

# Copy binaries
COPY --from=builder /app/server .
COPY --from=builder /app/seed .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
