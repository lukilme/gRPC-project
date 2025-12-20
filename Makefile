run-payment:
	@echo "Iniciando Payment Service..."
	cd microservices/payment && \
	DB_DRIVER=mysql \
	DATA_SOURCE_URL="root:password@tcp(127.0.0.1:3306)/payment" \
	APPLICATION_PORT=3001 \
	ENV=development \
	go run cmd/main.go

run-server:
	@echo "Iniciando Order Service...!"
	cd microservices/order && \
	DB_DRIVER=mysql \
	DATA_SOURCE_URL="root:password@tcp(127.0.0.1:3306)/order" \
	PAYMENT_SERVICE_URL="localhost:3001" \
	APPLICATION_PORT=3000 \
	ENV=development \
	go run cmd/main.go


test-db:
	@echo "Testando conexão com MySQL..."
	docker exec microservices-mysql mysql -h 127.0.0.1 -u root -ppassword -e "SHOW DATABASES;"

setup:
	@echo "Criando banco de dados..."
	docker exec microservices-mysql mysql -h 127.0.0.1 -u root -ppassword -e "CREATE DATABASE IF NOT EXISTS payment;"