package payment

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	paymentpb "ifpb.com/microservices-proto/golang/payment"
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
	client paymentpb.PaymentClient
	conn   *grpc.ClientConn
}

func NewAdapter(paymentServiceUrl string) (*Adapter, error) {
	retryOps := []grpc_retry.CallOption{
		grpc_retry.WithCodes(codes.Unavailable, codes.ResourceExhausted),
		grpc_retry.WithMax(5),
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(time.Second)),
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(retryOps...)),
	}

	log.Printf("%sConnecting to Payment Service: %s%s", colorBlue, paymentServiceUrl, colorReset)
	conn, err := grpc.NewClient(paymentServiceUrl, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	client := paymentpb.NewPaymentClient(conn)

	return &Adapter{
		client: client,
		conn:   conn,
	}, nil
}

func (a *Adapter) Close() {
	if a.conn != nil {
		a.conn.Close()
		log.Printf("%sConnection to Payment Service closed%s", colorBlue, colorReset)
	}
}

func (a *Adapter) ProcessPayment(ctx context.Context, order *domain.Order) (*domain.PaymentResponse, error) {
	log.Printf("%sProcessing payment for order %d, amount: %.2f%s",
		colorYellow, order.ID, order.TotalAmount, colorReset)

	req := &paymentpb.CreatePaymentRequest{
		CustomerId: order.CustomerID,
		OrderId:    order.ID,
		TotalPrice: float32(order.TotalAmount),
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := a.client.Create(ctx, req)
	if err != nil {
		log.Printf("%sPayment processing failed: %v%s", colorRed, err, colorReset)
		return nil, fmt.Errorf("payment processing failed: %w", err)
	}

	log.Printf("%sPayment processed: PaymentID=%d, BillID=%d%s",
		colorGreen, resp.GetPaymentId(), resp.GetBillId(), colorReset)

	status := "approved"
	if resp.GetPaymentId() == 0 {
		status = "failed"
	}

	return &domain.PaymentResponse{
		ID:     strconv.FormatInt(resp.GetPaymentId(), 10),
		Status: status,
		Amount: order.TotalAmount,
	}, nil
}
