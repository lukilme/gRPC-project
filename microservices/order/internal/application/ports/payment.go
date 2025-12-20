package ports

import "ifpb.com/microservices/order/internal/application/core/domain"

type PaymentPort interface {
	Charge(order *domain.Order) error
}
type OrderItem struct {
	ProductId int64
	Quantity  int32
	UnitPrice float32
}

type Order struct {
	ID         int64
	CustomerID int64
	OrderItems []OrderItem
}
