FROM golang:1.25f-alpine

WORKDIR /app

# Copiar arquivos de módulo
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar todo o código
COPY . .

# Build apontando para o diretório correto
RUN go build -o schema-registry ./cmd/app

# Expor porta
EXPOSE 8080

# Executar
CMD ["./schema-registry"]