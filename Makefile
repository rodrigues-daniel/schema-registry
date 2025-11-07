# Nome do binário
BINARY = schema-registry-bin

# Diretório principal do código
MAIN = ./cmd/server/main.go

# Variáveis de compilação
#
GOFLAGS = -ldflags="-s -w"

# Regra padrão
all: build

# Compilar o projeto
build:
	@echo " Building..."
	go build $(GOFLAGS) -o $(BINARY) $(MAIN)

# Executar o projeto
run:
	go run $(MAIN)

# Rodar os testes
test:
	@echo " Running tests..."
	go test ./... -v

# Formatar o código
fmt:
	go fmt ./...

# Rodar o linter (se tiver golangci-lint instalado)
lint:
	golangci-lint run ./...

# Limpar artefatos de build
clean:
	rm -f $(BINARY)

# Atualizar dependências
deps:
	go mod tidy
