package main

import (
	"log"

	_ "github.com/charmbracelet/bubbletea"
	"ifpb.com/microservices/order/config"
	"ifpb.com/microservices/order/internal/adapters/db"
	"ifpb.com/microservices/order/internal/adapters/grpc"
	payment_adapter "ifpb.com/microservices/order/internal/adapters/payment"
	"ifpb.com/microservices/order/internal/application/core/api"
)

func main() {
	log.Println("Iniciando microserviço Order...")

	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		log.Fatalf("Failed to connect to database. Error: %v", err)
	}
	defer func() {
	}()
	log.Println("Database adapter initialized")

	paymentAdapter, err := payment_adapter.NewAdapter(config.GetPaymentServiceUrl())
	if err != nil {
		log.Fatalf("Failed to initialize payment stub. Error: %v", err)
	}
	log.Println("Payment adapter initialized")

	application := api.NewApplication(dbAdapter, paymentAdapter)
	log.Println("Application core initialized")

	grpcAdapter := grpc.NewAdapter(application, config.GetApplicationPort())
	log.Printf("Starting gRPC server on port %s", config.GetApplicationPort())

	grpcAdapter.Run()
}
