package domain

type Order struct {
	ID         int64
	CustomerID int64
	Status     string
	OrderItems []OrderItem
	CreatedAt  int64
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

type PaymentPort interface {
	Charge(order *Order) error
}

type DBPort interface {
	Get(id int64) (Order, error)
	Save(order *Order) error
}
