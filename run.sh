#!/bin/bash

export DATA_SOURCE_URL="root:password@tcp(127.0.0.1:3306)/payment"
export APPLICATION_PORT="3001"
export ENV="development"

echo "iniciando Order Service..."
echo "database: $DATA_SOURCE_URL"
echo "port: $APPLICATION_PORT"

cd microservices/payment
go run cmd/main.go