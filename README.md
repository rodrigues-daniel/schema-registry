# 🧪 Schema Registry com NATS Embutido

> ⚠️ **Projeto Experimental:**  
> Esta aplicação é um protótipo em desenvolvimento que explora controle de versionamento e compatibilidade de *schemas* (Avro/JSON) utilizando **Golang** e um **servidor NATS embutido** diretamente no binário.  
> O objetivo é estudar abordagens leves e autônomas para registro de schemas, sem dependência de infraestrutura externa. Melhorias e novas funcionalidades estão **em andamento**.

---

## 🚀 Visão Geral

Um **Schema Registry completo** com:
- Servidor **NATS + JetStream embutido**
- API **RESTful**
- **Validação de compatibilidade** (forward, backward e full)
- **Observabilidade integrada** (métricas, logs e alertas)
- Suporte a **Docker e Kubernetes**

---

## 🧩 Características Principais

| Recurso | Descrição |
|----------|------------|
| 📝 **Schema Registry** | Armazenamento e versionamento de schemas Avro/JSON |
| ⚡ **NATS Embutido** | Servidor NATS + JetStream integrados (single binary) |
| 🔍 **Validação de Compatibilidade** | Forward, backward e full compatibility |
| 📊 **Observabilidade Completa** | Métricas, logs estruturados e alertas Prometheus/Grafana |
| 🐳 **Docker Ready** | Stack completa via Docker Compose |
| 🩺 **Health Checks** | Endpoints prontos para Kubernetes e balanceadores |

---

## 🏗️ Arquitetura

```
[Client] → [HTTP API :8080] → [Schema Registry] ↔ [NATS Embutido :4222]
                                      ↓
                [JetStream KV Store] → [Persistência em Arquivo]
```

---

## 📦 Quick Start

### 1️⃣ Clone e Build

```bash
git clone <seu-repositorio>
cd schema-registry

# se tiver o bin utils com make instalado
# Build da aplicação
make build

# Ou executar diretamente
make run
```

### 2️⃣ Docker Compose (Recomendado)

```bash
# Iniciar stack completa
make docker-up ou  docker compose up -d

# Verificar status
docker compose ps
```

### 3️⃣ Acessos

| Serviço | URL |
|----------|-----|
| API | [http://localhost:8080](http://localhost:8080) |
| Prometheus | [http://localhost:9090](http://localhost:9090) |
| Grafana | [http://localhost:3000](http://localhost:3000) — *(login: admin / admin)* |
| NATS Monitoring | [http://localhost:8222](http://localhost:8222) |

---

## 🔌 API Reference

### Health Checks

```bash
curl http://localhost:8080/health   # Básico
```

### Gerenciamento de Schemas

#### Registrar Schema
```bash
curl -X POST http://localhost:8080/schemas/user/versions   -H "Content-Type: application/json"   -d '{
    "schema": {
      "type": "record",
      "name": "User",
      "fields": [
        {"name": "id", "type": "int"},
        {"name": "name", "type": "string"},
        {"name": "email", "type": "string"}
      ]
    }
  }'
```

#### Recuperar Schema
```bash
curl http://localhost:8080/schemas/user/versions/1
curl http://localhost:8080/schemas/user/versions/latest
```

#### Listar Subjects e Versões
```bash
curl http://localhost:8080/subjects
curl http://localhost:8080/subjects/user/versions
```

---

## 🧪 Testes de Compatibilidade

O endpoint `/compatibility` permite testar evolução de schemas antes do registro.

```bash
curl -X POST http://localhost:8080/compatibility/subjects/user/versions   -H "Content-Type: application/json"   -d '{
    "schema": {
      "type": "record",
      "name": "User",
      "fields": [
        {"name": "id", "type": "int"},
        {"name": "name", "type": "string"},
        {"name": "email", "type": ["string", "null"]}
      ]
    }
  }'
```

Resposta esperada:
```json
{ "is_compatible": true }
```

---

## 🧰 Payloads para Testes (Postman)

Na pasta [`payloads/`](./payloads), você encontrará diversos arquivos JSON contendo **exemplos de requisições** para testar os endpoints do Schema Registry.

Esses arquivos podem ser **importados diretamente no Postman** para facilitar o envio de requisições.

### 👉 Como usar no Postman:
1. Abra o **Postman**.  
2. Clique em **Import** → **Upload Files**.  
3. Selecione os arquivos `.json` dentro da pasta `payloads/`.  
4. Execute as requisições conforme desejar (registro de schemas, compatibilidade, etc).

---

## 📊 Observabilidade

- **Métricas:** [http://localhost:8080/metrics](http://localhost:8080/metrics)  
- **Dashboards:** Grafana → Importar `monitoring/dashboard.json`

Métricas disponíveis:
- `schema_registry_registrations_total`
- `schema_registry_validations_total`
- `schema_registry_request_duration_seconds`
- `nats_jetstream_storage_bytes`

Alertas pré-configurados (Prometheus + Alertmanager):
- `SchemaRegistryDown`
- `HighErrorRate`
- `NATSConnectionIssues`

---

## 🔧 Configuração

### Variáveis de Ambiente

```bash
# NATS Embutido
NATS_SERVER_NAME=schema-registry-1
NATS_STORE_DIR=./jetstream-data
NATS_PORT=4222
NATS_HTTP_PORT=8222

# API HTTP
HTTP_PORT=:8080

# Observabilidade
METRICS_ENABLED=true
LOG_LEVEL=info
LOG_FORMAT=json

# Compatibilidade
COMPATIBILITY_LEVEL=BACKWARD
```

---

## 🐳 Docker

```bash
make docker-up       # Subir stack completa
make logs            # Logs em tempo real
make docker-down     # Parar e limpar
make docker-rebuild  # Rebuild completo
```

Serviços incluídos:
- `schema-registry` — Aplicação principal + NATS embutido  
- `prometheus` — Métricas  
- `grafana` — Dashboards  
- `alertmanager` — Alertas  
- `node-exporter` — Métricas do sistema  

---

## 📈 Monitoramento em Produção

### Health Checks para Kubernetes

```yaml
livenessProbe:
  httpGet:
    path: /live
    port: 8080
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
```

### Integração com Alertmanager (Slack)

```yaml
receivers:
  - name: 'slack-notifications'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/...'
        channel: '#alerts'
        title: 'Schema Registry Alert'
```

---

## 🧰 Status do Projeto

🔧 **Em andamento** — funcionalidades planejadas:
- [ ] Autenticação e ACLs
- [ ] Suporte a Protobuf
- [ ] Replicação distribuída entre instâncias
- [ ] UI web para gerenciamento de schemas

---

## 📜 Licença

Este projeto é disponibilizado sob a licença **MIT**.  
Sinta-se livre para testar, modificar e contribuir!
