package tui

import "fmt"

func (m Model) View() string {

	if m.Screen == MenuScreen {
		s := "MENU\n\n"

		for i, choice := range m.Menu {
			cursor := " "
			if m.Cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}

		s += "\n↑ ↓ navegar • enter selecionar • q sair\n"
		return s
	}

	return "NOVO PEDIDO\n\n" +
		"1-Customer ID\n" +
		"2-Product ID\n" +
		"3-Quantity ID\n" +
		"4-Unit Price\n" +
		m.CustomerID.View() + "\n" +
		m.ProductID.View() + "\n" +
		m.Quantity.View() + "\n" +
		m.UnitPrice.View() + "\n\n" +
		"enter avançar • esc voltar"
}
