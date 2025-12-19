package api

import (
	"ifpb.com/microservices/order/internal/application/core/domain"
)

type Application struct {
	db      domain.DBPort
	payment domain.PaymentPort
}

func NewApplication(db domain.DBPort, payment domain.PaymentPort) *Application {
	return &Application{
		db:      db,
		payment: payment,
	}
}

func (a Application) PlaceOrder(order domain.Order) (domain.Order, error) {
	err := a.db.Save(&order)
	if err != nil {
		return domain.Order{}, err
	}
	paymentErr := a.payment.Charge(&order)
	if paymentErr != nil {
		return domain.Order{}, paymentErr
	}
	return order, nil
}

func (a Application) GetOrder(id int64) (domain.Order, error) {
	return a.db.Get(id)
}
