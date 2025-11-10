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

# Cria usuário e diretórios COM a flag -u para definir UID/GID
RUN adduser -D -u 1000 appuser

# Cria diretório de dados com permissões corretas
RUN mkdir -p /data/jetstream && \
    chown -R appuser:appuser /data

WORKDIR /app

# Copia apenas o binário da etapa anterior
COPY --from=builder --chown=appuser:appuser /app/schema-registry .

# Usa o usuário não root
USER appuser

# Expõe TODAS as portas que sua aplicação usa
EXPOSE 8080 4222 8222

# Comando padrão
ENTRYPOINT ["./schema-registry"]