package api

import (
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ifpb.com/microservices/shipping/internal/application/core/domain"
)

const (
	blueColor   = "\033[34m"
	greenColor  = "\033[32m"
	yellowColor = "\033[33m"
	redColor    = "\033[31m"
	resetColor  = "\033[0m"
)

type Application struct {
	db domain.DBPort
}

func NewApplication(db domain.DBPort) *Application {
	log.Printf("%sINFO: Initializing Shipping Application with database port%s", blueColor, resetColor)
	return &Application{
		db: db,
	}
}

func (a *Application) CalculateAndPlaceShipping(orderID string, items []domain.OrderItem) (*domain.Shipping, error) {
	log.Printf("%sINFO: Starting shipping calculation for order: %s%s", blueColor, orderID, resetColor)
	log.Printf("%sDETAIL: Processing %d items for order %s%s", blueColor, len(items), orderID, resetColor)

	if len(items) == 0 {
		log.Printf("%sERROR: Cannot process shipping - no items provided for order %s%s", redColor, orderID, resetColor)
		return nil, status.Error(codes.InvalidArgument, "must have at least one item")
	}

	totalUnits := 0
	log.Printf("%sDETAIL: Validating item quantities...%s", blueColor, resetColor)

	for _, item := range items {
		if item.Quantity <= 0 {
			log.Printf("%sERROR: Invalid quantity for item %s: %d%s", redColor, item.ItemID, item.Quantity, resetColor)
			return &domain.Shipping{}, status.Errorf(codes.InvalidArgument, "quantity must be greater than zero")
		}
		totalUnits += item.Quantity
		log.Printf("%sDETAIL: Item %s: quantity %d (running total: %d)%s",
			blueColor, item.ItemID, item.Quantity, totalUnits, resetColor)
	}

	deliveryDays := 1
	if totalUnits > 0 {
		deliveryDays += totalUnits / 5
	}

	log.Printf("%sINFO: Delivery calculation complete - Total units: %d, Delivery days: %d%s",
		greenColor, totalUnits, deliveryDays, resetColor)

	shipping := &domain.Shipping{
		OrderID:      orderID,
		Items:        items,
		DeliveryDays: deliveryDays,
	}

	log.Printf("%sDETAIL: Attempting to save shipping to database...%s", blueColor, resetColor)
	if err := a.db.Save(shipping); err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("%sERROR: Database operation failed with status: %v - %s%s",
				redColor, st.Code(), st.Message(), resetColor)

			return &domain.Shipping{}, st.Err()
		}

		log.Printf("%sERROR: Failed to save shipping for order %s: %v%s", redColor, orderID, err, resetColor)

		return &domain.Shipping{}, status.Errorf(codes.Internal,
			"failed to save shipping for order %s: %v", orderID, err)
	}

	log.Printf("%sSUCCESS: Shipping calculated and saved for order %s: %d delivery days%s",
		greenColor, orderID, deliveryDays, resetColor)
	log.Printf("%sDETAIL: Shipping ID: %s, Items processed: %d%s",
		blueColor, shipping.ID, len(items), resetColor)

	return shipping, nil
}

func (a Application) PlaceShipping(shipping domain.Shipping) (domain.Shipping, error) {
	log.Printf("%sINFO: Placing shipping for order: %s%s", blueColor, shipping.OrderID, resetColor)
	log.Printf("%sDETAIL: Shipping details - ID: %s, Delivery days: %d, Items: %d%s",
		blueColor, shipping.ID, shipping.DeliveryDays, len(shipping.Items), resetColor)

	log.Printf("%sDETAIL: Saving shipping to database...%s", blueColor, resetColor)
	err := a.db.Save(&shipping)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("%sERROR: Failed to place shipping with status: %v - %s%s",
				redColor, st.Code(), st.Message(), resetColor)
			return domain.Shipping{}, st.Err()
		}

		log.Printf("%sERROR: Failed to place shipping for order %s: %v%s", redColor, shipping.OrderID, err, resetColor)
		return domain.Shipping{}, err
	}

	log.Printf("%sSUCCESS: Shipping placed successfully for order %s%s", greenColor, shipping.OrderID, resetColor)
	log.Printf("%sDETAIL: Shipping ID %s saved to database%s", blueColor, shipping.ID, resetColor)

	return shipping, nil
}

func (a *Application) GetShippingByOrderID(orderID int) (*domain.Shipping, error) {
	log.Printf("%sINFO: Retrieving shipping for order ID: %d%s", blueColor, orderID, resetColor)

	if orderID == 0 {
		log.Printf("%sERROR: Invalid order ID provided: %d%s", redColor, orderID, resetColor)
		return &domain.Shipping{}, status.Errorf(codes.InvalidArgument, "invalid order ID: must be greater than zero")
	}

	log.Printf("%sDETAIL: Querying database for order ID %d...%s", blueColor, orderID, resetColor)
	shipping, err := a.db.Get(orderID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("%sERROR: Database query failed with status: %v - %s%s",
				redColor, st.Code(), st.Message(), resetColor)
			return nil, st.Err()
		}

		log.Printf("%sERROR: Database error retrieving shipping for order %d: %v%s",
			redColor, orderID, err, resetColor)
		return nil, status.Errorf(codes.Internal, "failed to retrieve shipping: %v", err)
	}

	if shipping == nil {
		log.Printf("%sWARNING: No shipping found for order ID: %d%s", yellowColor, orderID, resetColor)
		return nil, errors.New("shipping not found")
	}

	log.Printf("%sSUCCESS: Shipping found for order %d%s", greenColor, orderID, resetColor)
	log.Printf("%sDETAIL: Shipping ID: %s, Delivery days: %d, Items count: %d%s",
		blueColor, shipping.ID, shipping.DeliveryDays, len(shipping.Items), resetColor)

	return shipping, nil
}
