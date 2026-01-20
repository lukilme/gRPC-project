package tui

import "github.com/charmbracelet/bubbles/textinput"

func InitForm() (textinput.Model, textinput.Model, textinput.Model, textinput.Model) {
	customer := textinput.New()
	customer.Placeholder = "Customer ID"
	customer.Focus()

	product := textinput.New()
	product.Placeholder = "Product ID"

	quantity := textinput.New()
	quantity.Placeholder = "Quantity"

	price := textinput.New()
	price.Placeholder = "Unit Price"

	return customer, product, quantity, price
}
