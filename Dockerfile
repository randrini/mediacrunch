# ── Build stage: Vue frontend ──────────────────────────────────────────────────
FROM node:20-alpine AS frontend-build

WORKDIR /app/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ .
RUN npm run build

# ── Build stage: Go backend ────────────────────────────────────────────────────
FROM golang:1.26.5-alpine AS backend-build

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

# Embed frontend build output
COPY --from=frontend-build /app/web/dist /app/web/dist

# Build arguments for version info (set by CI)
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X github.com/mediacrunch/mediacrunch/internal/version.Version=${VERSION} -X github.com/mediacrunch/mediacrunch/internal/version.Commit=${COMMIT} -X github.com/mediacrunch/mediacrunch/internal/version.BuildDate=${BUILD_DATE}" \
    -o /mediacrunch ./cmd/server/

# ── Runtime stage ──────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary
COPY --from=backend-build /mediacrunch /app/mediacrunch

# Copy frontend dist for serving (fallback if not embedded)
COPY --from=frontend-build /app/web/dist /app/web/dist

# Data directory for SQLite
RUN mkdir -p /app/data

ENV MC_DATA_DIR=/app/data
ENV MC_PORT=8080
ENV MC_QUALITY_DEFAULT=80
ENV MC_MAX_WIDTH_DEFAULT=1920
ENV MC_MIN_SAVING_KB=50

EXPOSE 8080

# Data volume
VOLUME ["/app/data"]

ENTRYPOINT ["/app/mediacrunch"]
