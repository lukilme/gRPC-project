package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	payment "ifpb.com/microservices-proto/golang/payment"
	"ifpb.com/microservices/order/internal/application/core/api"
	"ifpb.com/microservices/order/internal/application/core/domain"
)

type Adapter struct {
	payment.UnimplementedOrderServiceServer
	api    *api.Application
	port   string
	server *grpc.Server
}

func NewAdapter(api *api.Application, port string) *Adapter {
	return &Adapter{
		api:  api,
		port: port,
	}
}

func (a *Adapter) PlaceOrder(ctx context.Context, req *payment.CreateOrderRequest) (*payment.CreateOrderResponse, error) {
	order, err := convertToDomainOrder(req)
	if err != nil {
		return nil, err
	}

	savedOrder, err := a.api.PlaceOrder(order)
	if err != nil {
		return nil, err
	}

	return &payment.CreateOrderResponse{
		OrderId: savedOrder.ID,
	}, nil
}

func convertToDomainOrder(req *payment.CreateOrderRequest) (domain.Order, error) {
	var orderItems []domain.OrderItem
	for _, item := range req.GetItems() {
		orderItems = append(orderItems, domain.OrderItem{
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
			UnitPrice: item.GetUnitPrice(),
		})
	}

	return domain.Order{
		CustomerID: req.GetCustomerId(),
		OrderItems: orderItems,
	}, nil
}

func (a *Adapter) Run() {
	lis, err := net.Listen("tcp", ":"+a.port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	a.server = grpc.NewServer()

	payment.RegisterOrderServiceServer(a.server, a)

	reflection.Register(a.server)

	log.Printf("gRPC server listening on port %s", a.port)
	if err := a.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (a *Adapter) Stop() {
	if a.server != nil {
		a.server.GracefulStop()
	}
}
