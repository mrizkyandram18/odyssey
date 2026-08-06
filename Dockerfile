# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w \
      -X odyssey/pkg/observability.GitCommit=${GIT_COMMIT:-dev} \
      -X odyssey/pkg/observability.BuildDate=${BUILD_DATE:-dev} \
      -X odyssey/pkg/observability.Version=${VERSION:-dev} \
      -X odyssey/pkg/observability.SchemaVersion=${SCHEMA_VERSION:-12}" \
    -o /odyssey ./api/dev

# Frontend build stage
FROM node:20-alpine AS frontend
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Final stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /odyssey /odyssey
COPY --from=frontend /web/dist /app/web/dist

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD wget -q -O- "http://localhost:8080/health" || exit 1

ENTRYPOINT ["/odyssey"]
