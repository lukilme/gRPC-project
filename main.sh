#!/bin/bash

export DATA_SOURCE_URL="root:password@tcp(127.0.0.1:3306)/order"
export PAYMENT_SERVICE_URL="localhost:3001"
export APPLICATION_PORT="3000"
export ENV="development"

echo "iniciando Order Service..."
echo "database: $DATA_SOURCE_URL"
echo "payment Service: $PAYMENT_SERVICE_URL"
echo "port: $APPLICATION_PORT"

cd microservices/order
go run cmd/main.go