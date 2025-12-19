package grpc

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"ifpb.com/microservices/order/internal/application/core/api"
)

type OrderItem struct {
	ProductId int64
	Quantity  int32
	UnitPrice float32
}

type CreateOrderRequest struct {
	UserId int64
	Items  []*OrderItem
}

type CreateOrderResponse struct {
	OrderId int64
}

// Adicione este método ao struct Adapter
func (a *Adapter) PlaceOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// Sua lógica aqui usando a.api.PlaceOrder()
	return &CreateOrderResponse{OrderId: 999}, nil
}

type Adapter struct {
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

func (a *Adapter) Run() {
	lis, err := net.Listen("tcp", ":"+a.port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	a.server = grpc.NewServer()

	serviceDesc := &grpc.ServiceDesc{
		ServiceName: "order.OrderService",
		HandlerType: (*Adapter)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "PlaceOrder",
				Handler: func(srv interface{}, ctx context.Context,
					dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {

					var req CreateOrderRequest
					if err := dec(&req); err != nil {
						return nil, err
					}
					return srv.(*Adapter).PlaceOrder(ctx, &req)
				},
			},
		},
	}

	a.server.RegisterService(serviceDesc, a)
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
