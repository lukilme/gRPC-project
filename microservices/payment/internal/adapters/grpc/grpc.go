package grpc

import (
	"context"

	"github.com/huseyinbabal/microservices/payment/internal/application/core/domain"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	payment "ifpb.com/microservices-proto/golang/payment"
)

func (a Adapter) Create(
	ctx context.Context,
	request *payment.CreatePaymentRequest,
) (*payment.CreatePaymentResponse, error) {

	log.WithContext(ctx).Info("Creating payment...")
	newPayment := domain.NewPayment(
		request.CustomerId,
		request.OrderId,
		request.TotalPrice,
	)

	result, err := a.api.Charge(ctx, newPayment)
	if err != nil {
		code := status.Code(err)

		if code == codes.InvalidArgument {
			return nil, err
		}

		return nil, status.Errorf(
			codes.Internal,
			"failed to charge payment: %v",
			err,
		)
	}

	return &payment.CreatePaymentResponse{
		PaymentId: result.ID,
	}, nil
}
