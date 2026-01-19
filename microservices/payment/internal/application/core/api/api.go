package api

import (
	"context"

	"github.com/huseyinbabal/microservices/payment/internal/application/core/domain"
	"github.com/huseyinbabal/microservices/payment/internal/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db ports.DBPort
}

func NewApplication(db ports.DBPort) *Application {

	return &Application{
		db: db,
	}
}

func (a Application) Charge(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	if payment.TotalPrice > 1000 {
		payment.Status = "Canceled"
		_ = a.db.Save(ctx, &payment)
		return domain.Payment{}, status.Errorf(codes.InvalidArgument, "Payment over 1000 is not allowed.")
	}
	payment.Status = "Paid"

	if err := a.db.Save(ctx, &payment); err != nil {
		payment.Status = "Canceled"
		_ = a.db.Save(ctx, &payment)
		return domain.Payment{}, err
	}

	return payment, nil
}
