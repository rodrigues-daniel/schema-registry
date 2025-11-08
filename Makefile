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



.PHONY: up down build logs clean restart monitor test metrics

# Start all services
up:
	docker-compose up -d

# Start with build
up-build:
	docker-compose up -d --build

# Stop all services
down:
	docker-compose down

# Stop and remove volumes
down-clean:
	docker-compose down -v

# Build application
build:
	docker-compose build schema-registry

# View logs
logs:
	docker-compose logs -f schema-registry

# View all logs
logs-all:
	docker-compose logs -f

# Clean everything
clean:
	docker-compose down -v --remove-orphans
	rm -rf ./logs
	docker system prune -f

# Restart services
restart:
	docker-compose restart

# Monitor stack
monitor:
	@echo "📊 Monitoring Stack URLs:"
	@echo "  Schema Registry API: http://localhost:8080"
	@echo "  Prometheus:         http://localhost:9090"
	@echo "  Grafana:            http://localhost:3000 (admin/admin123)"
	@echo "  Alertmanager:       http://localhost:9093"
	@echo "  NATS Monitoring:    http://localhost:8222"
	@echo "  Jaeger UI:          http://localhost:16686"
	@echo "  Portainer:          http://localhost:9000"

# Run tests
test:
	docker-compose run --rm test-client

# Check metrics
metrics:
	curl -s http://localhost:8080/metrics | head -20

# Health check
health:
	curl http://localhost:8080/health | jq .

# Scale schema registry
scale:
	docker-compose up -d --scale schema-registry=3

# Backup data
backup:
	mkdir -p backups
	tar -czf backups/backup-$(shell date +%Y%m%d-%H%M%S).tar.gz volumes/

# Performance test
perf-test:
	docker run --rm --network schema-registry_monitoring \
		williamyeh/wrk -t4 -c100 -d30s http://schema-registry:8080/health

# Status
status:
	docker-compose ps
	docker-compose images