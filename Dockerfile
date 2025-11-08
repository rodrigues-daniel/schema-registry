# ---------------------------
# 🏗️ Etapa 1: Build
# ---------------------------
FROM golang:1.25 AS builder

WORKDIR /app

# Copia e baixa dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código
COPY . .

# Compila binário estaticamente
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o schema-registry ./cmd/server

# ---------------------------
# 📦 Etapa 2: Runtime leve
# ---------------------------
FROM alpine:3.20

# Adiciona um usuário não root (boa prática de segurança)
RUN adduser -D appuser

WORKDIR /app

# Copia apenas o binário da etapa anterior
COPY --from=builder /app/schema-registry .

# Usa o usuário não root
USER appuser

EXPOSE 8080

# Comando padrão
ENTRYPOINT ["./schema-registry"]