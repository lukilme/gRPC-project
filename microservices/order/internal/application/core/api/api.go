package api

import (
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	total_item := 0
	for _, item := range order.OrderItems {
		total_item += int(item.Quantity)
	}
	if total_item > 50 {
		return domain.Order{}, status.Errorf(codes.ResourceExhausted, "Order cannot over 50 items.")
	}
	err := a.db.Save(&order)
	if err != nil {
		return domain.Order{}, err
	}
	totalPrice := order.TotalPrice()
	log.Println(totalPrice)
	paymentErr := a.payment.Charge(&order)
	if paymentErr != nil {
		return domain.Order{}, paymentErr
	}
	return order, nil
}

func (a Application) GetOrder(id int64) (domain.Order, error) {
	order, err := a.db.Get(id)
	if err != nil {
		return domain.Order{}, err
	}
	return order, nil
}
