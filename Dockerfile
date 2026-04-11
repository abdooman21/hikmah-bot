# ── Stage 1: Build ──────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

# git is needed for go mod download with VCS deps
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o hikmah-bot ./main.go

# ── Stage 2: Runtime ────────────────────────────────────────────
FROM alpine:3.20

# ffmpeg: for voice/audio streaming
# ca-certificates: for HTTPS connections (radio URL, Discord API)
RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /app

# Copy the compiled binary from the build stage
COPY --from=builder /app/hikmah-bot .

# Copy the icons folder your code references at runtime
COPY --from=builder /app/internal/database/icons ./internal/database/icons

CMD ["./hikmah-bot"]