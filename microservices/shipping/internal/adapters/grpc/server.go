package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	shippingpb "ifpb.com/microservices-proto/golang/shipping"
	"ifpb.com/microservices/shipping/internal/application/core/api"
	"ifpb.com/microservices/shipping/internal/application/core/domain"
)

type Adapter struct {
	shippingpb.UnimplementedShippingServiceServer
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

func (a *Adapter) CalculateShipping(ctx context.Context, req *shippingpb.ShippingRequest) (*shippingpb.ShippingResponse, error) {
	log.Printf("[Shipping] Recebida requisição para pedido: %s", req.GetOrderId())

	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id é obrigatório")
	}

	if len(req.GetItems()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "é necessário pelo menos um item")
	}

	domainItems := make([]domain.OrderItem, len(req.GetItems()))
	for i, pbItem := range req.GetItems() {
		if pbItem.GetQuantity() <= 0 {
			return nil, status.Errorf(codes.InvalidArgument, "quantidade inválida para item %s", pbItem.GetItemId())
		}

		domainItems[i] = domain.OrderItem{
			ItemID:   pbItem.GetItemId(),
			Quantity: int(pbItem.GetQuantity()),
		}
	}
	value, err := strconv.Atoi(req.GetOrderId())
	if err != nil {
		log.Printf("gRPC Error em %s: %v", "dasd", err)
	}
	shipping, err := a.api.CalculateAndPlaceShipping(value, domainItems)

	if err != nil {
		log.Printf("[Shipping] Erro ao processar shipping: %v", err)

		if err.Error() == "orderID não pode ser vazio" ||
			err.Error() == "deve haver pelo menos um item" ||
			err.Error() == "quantidade deve ser maior que zero" {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, "erro interno ao processar shipping")
	}

	log.Printf("[Shipping] Shipping calculado para pedido %s: %d dias",
		shipping.OrderID, shipping.DeliveryDays)

	str := strconv.Itoa(shipping.OrderID)
	return &shippingpb.ShippingResponse{
		OrderId:      str,
		DeliveryDays: int32(shipping.DeliveryDays),
		Status:       "CALCULATED",
	}, nil
}

func (a *Adapter) Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", a.port, err)
	}

	a.server = grpc.NewServer(
		grpc.UnaryInterceptor(a.loggingInterceptor()),
	)

	shippingpb.RegisterShippingServiceServer(a.server, a)

	reflection.Register(a.server)

	log.Printf("Shipping Service gRPC server escutando na porta %d", a.port)

	if err := a.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

func (a *Adapter) Stop() {
	log.Println("Parando servidor gRPC do Shipping Service...")
	if a.server != nil {
		a.server.GracefulStop()
		log.Println("Servidor gRPC parado com sucesso")
	}
}

func (a *Adapter) loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod
		log.Printf("📨 gRPC Request: %s", method)

		if method == "/shipping.ShippingService/CalculateShipping" {
			if shippingReq, ok := req.(*shippingpb.ShippingRequest); ok {
				log.Printf("   Pedido: %s, Itens: %d",
					shippingReq.GetOrderId(), len(shippingReq.GetItems()))
			}
		}

		resp, err := handler(ctx, req)

		if err != nil {
			log.Printf("gRPC Error em %s: %v", method, err)
		} else {
			log.Printf("gRPC Success em %s", method)
		}

		return resp, err
	}
}
