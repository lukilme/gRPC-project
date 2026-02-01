package api

import (
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ifpb.com/microservices/shipping/internal/application/core/domain"
)

type Application struct {
	db domain.DBPort
}

func NewApplication(db domain.DBPort, shipping domain.DBPort) *Application {
	return &Application{
		db: db,
	}
}

func (a *Application) CalculateAndPlaceShipping(orderID int, items []domain.OrderItem) (*domain.Shipping, error) {

	if len(items) == 0 {
		return nil, errors.New("deve haver pelo menos um item")
	}

	totalUnits := 0
	for _, item := range items {
		if item.Quantity <= 0 {
			// return nil, errors.New("quantidade deve ser maior que zero")
			return &domain.Shipping{}, status.Errorf(codes.InvalidArgument, "Quantidade deve ser maior que zero")
		}
		totalUnits += item.Quantity
	}

	deliveryDays := 1
	if totalUnits > 0 {
		deliveryDays += totalUnits / 5
	}

	shipping := &domain.Shipping{
		OrderID:      orderID,
		Items:        items,
		DeliveryDays: deliveryDays,
	}

	if err := a.db.Save(shipping); err != nil {
		log.Printf("Erro ao salvar shipping: %v", err)
		// return nil, errors.New("falha ao processar shipping")
		return &domain.Shipping{}, status.Errorf(codes.Internal, "falha ao processar shipping")
	}

	log.Printf("Shipping calculado e salvo para pedido %s: %d dias", orderID, deliveryDays)
	return shipping, nil
}

func (a Application) PlaceShipping(shipping domain.Shipping) (domain.Shipping, error) {

	err := a.db.Save(&shipping)
	if err != nil {
		return domain.Shipping{}, err
	}
	return shipping, nil
}

func (a *Application) GetShippingByOrderID(orderID int) (*domain.Shipping, error) {
	if orderID == 0 {
		return &domain.Shipping{}, status.Errorf(codes.InvalidArgument, "número de id inválido")
	}

	shipping, err := a.db.Get(orderID)
	if err != nil {
		log.Printf("Erro ao buscar shipping: %v", err)
		return nil, errors.New("falha ao buscar shipping")
	}

	if shipping == nil {
		return nil, errors.New("shipping não encontrado")
	}

	return shipping, nil
}
