FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /gopulse ./cmd/server/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS calls if needed
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /gopulse /app/gopulse
COPY migrations /app/migrations
COPY frontend /app/frontend

# Expose port
EXPOSE 8080

CMD ["/app/gopulse"]
