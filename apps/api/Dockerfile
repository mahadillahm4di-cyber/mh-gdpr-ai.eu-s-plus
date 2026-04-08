# ── Stage 1: Build ──
FROM golang:1.22-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mh-gdpr-api ./cmd/server

# ── Stage 2: Runtime ──
# SECURITY: Minimal image, non-root user, read-only filesystem
FROM alpine:3.19

RUN apk add --no-cache ca-certificates wget \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /build/mh-gdpr-api .

RUN mkdir -p /app/data && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget --spider -q http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["./mh-gdpr-api"]
