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
	var total float32
	for _, item := range o.OrderItems {
		total += item.UnitPrice * float32(item.Quantity)
	}
	return total
}
