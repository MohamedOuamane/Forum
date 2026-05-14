# =========================
# Étape 1 — Build
# =========================
FROM golang:1.25 AS builder

WORKDIR /app

# CGO nécessaire pour sqlite3
RUN apt-get update && apt-get install -y gcc libc6-dev

# Dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le projet
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o forum .

# =========================
# Étape 2 — Runtime
# =========================
FROM debian:bookworm-slim

WORKDIR /app

# Runtime minimal pour SQLite
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Copier le binaire
COPY --from=builder /app/forum .

# Copier assets
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/STATIC ./STATIC
COPY --from=builder /app/database ./database

# Copier .env
COPY --from=builder /app/.env ./.env

# Port
EXPOSE 8080

# Lancer l'app
CMD ["./forum"]