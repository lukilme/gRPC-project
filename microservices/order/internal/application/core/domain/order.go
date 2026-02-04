package domain

import "context"

type Order struct {
	ID           int64
	CustomerID   int64
	Status       string
	OrderItems   []OrderItem
	CreatedAt    int64
	TotalAmount  float32
	DeliveryDays int
}

type OrderItem struct {
	ProductID int64
	Quantity  int32
	UnitPrice float32
}

func (o *Order) TotalPrice() float32 {
	var totalPrice float32
	for _, orderItem := range o.OrderItems {
		totalPrice += orderItem.UnitPrice * float32(orderItem.Quantity)
	}
	return totalPrice
}

type DBPort interface {
	Get(id int64) (Order, error)
	Save(order *Order) error
	Update(order *Order) error
	ItemExists(productID int64) (bool, error)
}

type PaymentClient interface {
	ProcessPayment(ctx context.Context, order *Order) (*PaymentResponse, error)
}

type ShippingClient interface {
	CalculateShipping(ctx context.Context, orderID string, items []ShippingItem) (*ShippingResponse, error)
	Close()
}

type PaymentResponse struct {
	ID     string
	Status string
	Amount float32
}

type ShippingResponse struct {
	OrderID      string
	DeliveryDays int
	Status       string
}

type ShippingItem struct {
	ItemID   string
	Quantity int32
}
