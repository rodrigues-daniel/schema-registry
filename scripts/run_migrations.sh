#!/bin/bash

# Carregar variáveis de ambiente
set -a
source ../.env
set +a

# Construir string de conexão
DB_STRING="host=${DB_HOST} port=${DB_PORT} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} sslmode=${DB_SSL_MODE}"

echo "Iniciando migrações do banco de dados..."
echo "Database: ${DB_NAME}"
echo "Host: ${DB_HOST}:${DB_PORT}"

# Função para executar comandos goose
run_goose() {
    goose -dir "${MIGRATIONS_DIR}" postgres "${DB_STRING}" "$@"
}

case "$1" in
    "status")
        run_goose status
        ;;
    "up")
        run_goose up
        ;;
    "down")
        run_goose down
        ;;
    "redo")
        run_goose redo
        ;;
    "reset")
        run_goose reset
        ;;
    "create")
        if [ -z "$2" ]; then
            echo "Uso: $0 create <nome_da_migration>"
            exit 1
        fi
        goose -dir "${MIGRATIONS_DIR}" create "$2" sql
        ;;
    *)
        echo "Uso: $0 {status|up|down|redo|reset|create}"
        echo ""
        echo "Comandos:"
        echo "  status    Ver status das migrações"
        echo "  up        Aplicar todas as migrações"
        echo "  down      Reverter uma migração"
        echo "  redo      Re-executar última migração"
        echo "  reset     Reverter todas as migrações"
        echo "  create    Criar nova migração"
        exit 1
esac