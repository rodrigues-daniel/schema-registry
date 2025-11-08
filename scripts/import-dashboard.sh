#!/bin/bash
# scripts/import-dashboard.sh

echo " Importando dashboard para Grafana..."

# Aguardar Grafana estar pronto
until curl -s http://localhost:3000/api/health; do
    echo "Aguardando Grafana..."
    sleep 5
done

# Importar dashboard via API
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d @../monitoring/dashboard.json \
  "http://admin:admin123@localhost:3000/api/dashboards/import"

echo "Dashboard importado com sucesso!"
echo "Acesse: http://localhost:3000"