package api

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"ifpb.com/microservices/order/internal/application/core/domain"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

type Application struct {
	db             domain.DBPort
	paymentClient  domain.PaymentClient
	shippingClient domain.ShippingClient
}

func NewApplication(db domain.DBPort, paymentClient domain.PaymentClient, shippingClient domain.ShippingClient) *Application {
	return &Application{
		db:             db,
		paymentClient:  paymentClient,
		shippingClient: shippingClient,
	}
}

func (a *Application) PlaceOrder(ctx context.Context, order domain.Order) (*domain.Order, error) {
	log.Printf("%sProcessing order for customer: %d%s", colorYellow, order.CustomerID, colorReset)

	for _, item := range order.OrderItems {
		exists, err := a.db.ItemExists(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("error checking item: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("item with ID %d not found in stock", item.ProductID)
		}
	}

	order.TotalAmount = order.TotalPrice()
	order.Status = "pending"

	err := a.db.Save(&order)
	if err != nil {
		return nil, fmt.Errorf("error saving order: %w", err)
	}

	log.Printf("%sOrder %d saved. Processing payment...%s", colorBlue, order.ID, colorReset)

	paymentResp, err := a.paymentClient.ProcessPayment(ctx, &order)
	if err != nil {
		order.Status = "payment_failed"
		a.db.Update(&order)
		return nil, fmt.Errorf("payment error: %w", err)
	}

	if paymentResp.Status == "approved" {
		log.Printf("%sPayment approved. Calculating shipping...%s", colorGreen, colorReset)

		shippingItems := make([]domain.ShippingItem, len(order.OrderItems))
		for i, item := range order.OrderItems {
			shippingItems[i] = domain.ShippingItem{
				ItemID:   strconv.FormatInt(item.ProductID, 10),
				Quantity: item.Quantity,
			}
		}

		orderIDStr := strconv.FormatInt(order.ID, 10)

		shippingResp, err := a.shippingClient.CalculateShipping(ctx, orderIDStr, shippingItems)
		if err != nil {
			log.Printf("%sError calculating shipping: %v%s", colorRed, err, colorReset)
			order.Status = "shipping_error"
		} else {
			order.DeliveryDays = shippingResp.DeliveryDays
			order.Status = "completed"
			log.Printf("%sShipping calculated: %d days%s", colorGreen, order.DeliveryDays, colorReset)
		}
	} else {
		order.Status = "payment_failed"
	}

	err = a.db.Update(&order)
	if err != nil {
		return nil, fmt.Errorf("error updating order: %w", err)
	}

	log.Printf("%sOrder %d completed with status: %s%s",
		colorGreen, order.ID, order.Status, colorReset)

	return &order, nil
}
