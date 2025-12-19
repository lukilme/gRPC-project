package main

import (
	"os"
	"testing"

	"ifpb.com/microservices/order/internal/application/core/domain"
)

func TestMain(m *testing.M) {
	// Configurar ambiente de teste
	os.Setenv("DATA_SOURCE_URL", "root:minhasenha@tcp(127.0.0.1:3306)/order_test")
	os.Setenv("PAYMENT_SERVICE_URL", "localhost:3001")

	code := m.Run()
	os.Exit(code)
}

func TestPlaceOrder(t *testing.T) {
	order := domain.Order{
		CustomerID: 123,
		Status:     "pending",
		OrderItems: []domain.OrderItem{
			{ProductID: 1, Quantity: 2, UnitPrice: 10.5},
		},
	}

	total := order.TotalPrice()
	if total != 21.0 {
		t.Errorf("TotalPrice() = %v, want %v", total, 21.0)
	}
}
