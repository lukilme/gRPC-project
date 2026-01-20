package grpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"ifpb.com/client-tui/domain"
	pb "ifpb.com/microservices-proto/golang/payment"
)

func PlaceOrder(order domain.Order) error {
	conn, err := grpc.Dial("localhost:3000", grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewOrderServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = client.PlaceOrder(ctx, &pb.CreateOrderRequest{
		CustomerId: order.CustomerID,
		Items: []*pb.OrderItem{
			{
				ProductId: order.OrderItems[0].ProductID,
				Quantity:  order.OrderItems[0].Quantity,
				UnitPrice: order.OrderItems[0].UnitPrice,
			},
		},
	})

	return err
}
