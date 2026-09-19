# =============================================================================
# Stage 1: Build — Go 1.27 + Node (for Tailwind CSS via pnpm)
# =============================================================================
FROM golang:1.27-alpine AS builder

# Install Node.js, pnpm, and build tools
RUN apk add --no-cache nodejs npm git \
    && npm install -g pnpm

WORKDIR /app

# Copy dependency manifests first (layer-cache friendly)
COPY go.mod go.sum ./
RUN go mod download

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

# Copy the rest of the source
COPY . .

# 1. Build the WASM client bundle
RUN GOARCH=wasm GOOS=js go build -o web/app.wasm

# 2. Compile Tailwind CSS
RUN pnpm exec tailwindcss -i styles/main.css -o web/styles.css

# 3. Build the server binary (statically linked for Alpine)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/task-timer .

# =============================================================================
# Stage 2: Runtime — minimal Alpine image
# =============================================================================
FROM alpine:3.22

# ca-certificates is needed for outbound HTTPS (e.g. go-app CDN assets)
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the server binary
COPY --from=builder /app/bin/task-timer ./task-timer

# Copy static web assets (WASM + CSS) served by go-app Handler
COPY --from=builder /app/web ./web

EXPOSE 8080

ENTRYPOINT ["./task-timer"]
