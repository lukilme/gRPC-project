package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
)

type Screen int

const (
	MenuScreen Screen = iota
	FormScreen
)

type Model struct {
	Screen Screen
	Cursor int

	Menu []string

	CustomerID textinput.Model
	ProductID  textinput.Model
	Quantity   textinput.Model
	UnitPrice  textinput.Model

	FocusIndex int
	Err        error
}

func InitialModel() Model {
	customer, product, qty, price := InitForm()

	return Model{
		Screen:     MenuScreen,
		Cursor:     0,
		Menu:       []string{"Novo Pedido", "Sair"},
		CustomerID: customer,
		ProductID:  product,
		Quantity:   qty,
		UnitPrice:  price,
		FocusIndex: 0,
	}
}
