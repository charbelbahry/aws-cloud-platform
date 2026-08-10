# Stage 1: Build static binary
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache Go modules layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build minimal static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api

# Stage 2: Production Minimal Runtime
FROM alpine:3.24 AS runtime

RUN apk --no-cache add ca-certificates tzdata \
    && addgroup -S appgroup -g 1000 \
    && adduser -S appuser -u 1000 -G appgroup

WORKDIR /app

COPY --from=builder /api /api

USER 1000:1000

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/api"]
