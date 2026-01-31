package domain

type Shipping struct {
	OrderID      int
	Items        []OrderItem
	DeliveryDays int
}

type OrderItem struct {
	ItemID   string
	Quantity int
}

type ShippingRepository interface {
	CalculateDeliveryDays(items []OrderItem) (int, error)
}
