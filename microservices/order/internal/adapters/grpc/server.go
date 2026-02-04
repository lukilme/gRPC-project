package grpc

import (
	"context"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	orderpb "ifpb.com/microservices-proto/golang/order"
	"ifpb.com/microservices/order/internal/application/core/api"
	"ifpb.com/microservices/order/internal/application/core/domain"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

type Adapter struct {
	orderpb.UnimplementedOrderServiceServer
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

func (a *Adapter) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	log.Printf("%sReceived order from customer: %s%s", colorBlue, req.GetCustomerId(), colorReset)

	customerID, err := strconv.ParseInt(req.GetCustomerId(), 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid customer_id: %v", err)
	}

	domainOrder, err := convertToDomainOrder(req, customerID)
	if err != nil {
		return nil, err
	}

	savedOrder, err := a.api.PlaceOrder(ctx, domainOrder)
	if err != nil {
		log.Printf("%sError processing order: %v%s", colorRed, err, colorReset)
		return nil, status.Errorf(codes.Internal, "error processing order: %v", err)
	}

	log.Printf("%sOrder processed successfully: ID=%d, Total=%.2f%s",
		colorGreen, savedOrder.ID, savedOrder.TotalAmount, colorReset)

	return convertToGRPCResponse(savedOrder), nil
}

func convertToDomainOrder(req *orderpb.CreateOrderRequest, customerID int64) (domain.Order, error) {
	var orderItems []domain.OrderItem

	for _, item := range req.GetItems() {
		productID, err := strconv.ParseInt(item.GetProductId(), 10, 64)
		if err != nil {
			return domain.Order{}, status.Errorf(codes.InvalidArgument,
				"invalid product_id: %s", item.GetProductId())
		}

		if item.GetQuantity() <= 0 {
			return domain.Order{}, status.Errorf(codes.InvalidArgument,
				"invalid quantity for product %s", item.GetProductId())
		}

		orderItems = append(orderItems, domain.OrderItem{
			ProductID: productID,
			Quantity:  item.GetQuantity(),
			UnitPrice: float32(item.GetUnitPrice()),
		})
	}

	return domain.Order{
		CustomerID: customerID,
		OrderItems: orderItems,
		Status:     "pending",
	}, nil
}

func convertToGRPCResponse(order *domain.Order) *orderpb.CreateOrderResponse {
	response := &orderpb.CreateOrderResponse{
		OrderId:      strconv.FormatInt(order.ID, 10),
		CustomerId:   strconv.FormatInt(order.CustomerID, 10),
		TotalAmount:  float64(order.TotalAmount),
		Status:       order.Status,
		DeliveryDays: int32(order.DeliveryDays),
	}

	for _, item := range order.OrderItems {
		response.Items = append(response.Items, &orderpb.OrderItem{
			ProductId: strconv.FormatInt(item.ProductID, 10),
			Quantity:  item.Quantity,
			UnitPrice: float64(item.UnitPrice),
		})
	}

	return response
}

func (a *Adapter) Run() {
	lis, err := net.Listen("tcp", ":"+a.port)
	if err != nil {
		log.Fatalf("%sfailed to listen: %v%s", colorRed, err, colorReset)
	}

	a.server = grpc.NewServer(
		grpc.UnaryInterceptor(a.loggingInterceptor()),
	)

	orderpb.RegisterOrderServiceServer(a.server, a)

	reflection.Register(a.server)

	log.Printf("%sOrder Service gRPC server listening on port %s%s", colorGreen, a.port, colorReset)
	if err := a.server.Serve(lis); err != nil {
		log.Fatalf("%sfailed to serve: %v%s", colorRed, err, colorReset)
	}
}

func (a *Adapter) Stop() {
	log.Printf("%sStopping Order Service...%s", colorYellow, colorReset)
	if a.server != nil {
		a.server.GracefulStop()
	}
}

func (a *Adapter) loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Printf("%sOrder Service - Method: %s%s", colorBlue, info.FullMethod, colorReset)
		return handler(ctx, req)
	}
}
