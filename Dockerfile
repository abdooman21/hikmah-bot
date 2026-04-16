# ── Stage 1: Build ──────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -o hikmah-bot ./main.go

# ── Stage 2: Runtime ────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ffmpeg ca-certificates

WORKDIR /app

COPY --from=builder /app/hikmah-bot .

COPY --from=builder /app/internal/database/icons ./internal/database/icons

CMD ["./hikmah-bot"]