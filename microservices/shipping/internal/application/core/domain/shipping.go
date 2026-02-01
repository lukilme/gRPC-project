package domain

type Shipping struct {
	ID           int
	OrderID      int
	Items        []OrderItem
	DeliveryDays int
	CreatedAt    int
}

type OrderItem struct {
	ItemID   string
	Quantity int
}

type DBPort interface {
	Save(shipping *Shipping) error
	Get(orderID int) (*Shipping, error)
	GetByID(shippingID int) (*Shipping, error)
	Update(shipping *Shipping) error
	CalculateDeliveryDays(items []OrderItem) (int, error)
}
