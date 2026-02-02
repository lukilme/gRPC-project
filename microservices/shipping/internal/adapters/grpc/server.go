package grpc

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"ifpb.com/microservices-proto/golang/payment"
	"ifpb.com/microservices/shipping/internal/application/core/api"
	"ifpb.com/microservices/shipping/internal/application/core/domain"
)

type Adapter struct {
	payment.UnimplementedOrderServiceServer
	api    *api.Application
	port   int
	server *grpc.Server
}

func NewAdapter(api *api.Application, port int) *Adapter {
	return &Adapter{
		api:  api,
		port: port,
	}
}

type ShippingService struct {
	db domain.DBPort
}

func (s *ShippingService) CalculateShipping(ctx context.Context, orderID int, items []domain.OrderItem) (*domain.Shipping, error) {
	deliveryDays, err := s.db.CalculateDeliveryDays(items)
	if err != nil {
		return nil, err
	}

	// Regra de negócio: mínimo 1 dia + 1 dia a cada 5 unidades
	totalUnits := 0
	for _, item := range items {
		totalUnits += item.Quantity
	}

	deliveryDays = int(math.Max(1, float64(1+totalUnits/5)))

	shipping := &domain.Shipping{
		OrderID:      orderID,
		Items:        items,
		DeliveryDays: deliveryDays,
	}

	return shipping, nil
}

func (a *Adapter) Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	a.server = grpc.NewServer()

	payment.RegisterOrderServiceServer(a.server, a)

	reflection.Register(a.server)

	log.Printf("gRPC server listening on port %v", a.port)
	if err := a.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (a *Adapter) Stop() {
	if a.server != nil {
		a.server.GracefulStop()
	}
}
