# Builder stage
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Accept build arguments for versioning
ARG VERSION=1.0.0
ARG COMMIT_HASH=unknown
ARG BUILD_TIME=unknown

COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .

# Build binary with injected ldflags
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w -X main.Version=${VERSION} -X main.CommitHash=${COMMIT_HASH} -X main.BuildTime=${BUILD_TIME}" \
  -o /server ./cmd/server

# Runner stage
FROM alpine:latest
WORKDIR /root/
ENV DOCS_DIR=/app/docs

RUN apk add --no-cache curl

COPY --from=builder /server .
COPY docs/api /app/docs/api

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health/ready || exit 1

EXPOSE 8080
CMD ["./server"]