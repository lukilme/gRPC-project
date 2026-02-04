package main

import (
	"log"

	"ifpb.com/microservices/order/config"
	"ifpb.com/microservices/order/internal/adapters/db"
	"ifpb.com/microservices/order/internal/adapters/grpc"
	payment_adapter "ifpb.com/microservices/order/internal/adapters/payment"
	"ifpb.com/microservices/order/internal/adapters/shipping"
	"ifpb.com/microservices/order/internal/application/core/api"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

func main() {
	log.Printf("%sStarting Order microservice...%s", colorGreen, colorReset)

	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		log.Fatalf("%sFailed to connect to database. Error: %v%s", colorRed, err, colorReset)
	}
	defer func() {
		if err := dbAdapter.Close(); err != nil {
			log.Printf("%sError closing database connection: %v%s", colorRed, err, colorReset)
		}
	}()
	log.Printf("%sDatabase adapter initialized%s", colorGreen, colorReset)

	paymentAdapter, err := payment_adapter.NewAdapter(config.GetPaymentServiceUrl())
	if err != nil {
		log.Fatalf("%sFailed to initialize payment stub. Error: %v%s", colorRed, err, colorReset)
	}
	defer func() {
		if paymentAdapter != nil {
			paymentAdapter.Close()
		}
	}()
	log.Printf("%sPayment adapter initialized%s", colorGreen, colorReset)

	shippingAdapter, err := shipping.NewShippingClient(config.GetShippingServiceUrl())
	if err != nil {
		log.Fatalf("%sFailed to initialize shipping client. Error: %v%s", colorRed, err, colorReset)
	}
	defer func() {
		if shippingAdapter != nil {
			shippingAdapter.Close()
		}
	}()
	log.Printf("%sShipping adapter initialized%s", colorGreen, colorReset)

	application := api.NewApplication(dbAdapter, paymentAdapter, shippingAdapter)
	log.Printf("%sApplication core initialized%s", colorGreen, colorReset)

	grpcAdapter := grpc.NewAdapter(application, config.GetApplicationPort())
	log.Printf("%sStarting gRPC server on port %s%s", colorGreen, config.GetApplicationPort(), colorReset)

	log.Printf("%sChecking connectivity with dependent services...%s", colorBlue, colorReset)

	grpcAdapter.Run()
}
