# Build stage
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO (required for sqlite3)
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o cardsplit ./cmd/cardsplit/main.go

# Final stage
FROM alpine:latest
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/cardsplit .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/web ./web

EXPOSE 8080
CMD ["./cardsplit"]
