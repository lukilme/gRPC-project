package shipping

import (
	"context"
	"log"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	shippingpb "ifpb.com/microservices-proto/golang/shipping"
	"ifpb.com/microservices/order/internal/application/core/domain"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

type ShippingClient struct {
	client shippingpb.ShippingServiceClient
	conn   *grpc.ClientConn
}

func NewShippingClient(shippingServiceUrl string) (*ShippingClient, error) {
	retryOps := []grpc_retry.CallOption{
		grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
		grpc_retry.WithMax(5),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOps...)),
	}

	log.Printf("%sConnecting to Shipping Service: %s%s", colorBlue, shippingServiceUrl, colorReset)
	conn, err := grpc.NewClient(shippingServiceUrl, opts...)
	if err != nil {
		return nil, err
	}

	client := shippingpb.NewShippingServiceClient(conn)

	return &ShippingClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *ShippingClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *ShippingClient) CalculateShipping(ctx context.Context, orderID string, items []domain.ShippingItem) (*domain.ShippingResponse, error) {
	pbItems := make([]*shippingpb.OrderItem, len(items))
	for i, item := range items {
		pbItems[i] = &shippingpb.OrderItem{
			ItemId:   item.ItemID,
			Quantity: item.Quantity,
		}
	}

	req := &shippingpb.ShippingRequest{
		OrderId: orderID,
		Items:   pbItems,
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.client.CalculateShipping(ctx, req)
	if err != nil {
		log.Printf("%sShipping calculation failed: %v%s", colorRed, err, colorReset)
		return nil, err
	}

	log.Printf("%sShipping calculated: OrderID=%s, Days=%d, Status=%s%s",
		colorGreen, resp.GetOrderId(), resp.GetDeliveryDays(), resp.GetStatus(), colorReset)

	return &domain.ShippingResponse{
		OrderID:      resp.GetOrderId(),
		DeliveryDays: int(resp.GetDeliveryDays()),
		Status:       resp.GetStatus(),
	}, nil
}
