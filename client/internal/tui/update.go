package tui

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"ifpb.com/client-tui/domain"
	"ifpb.com/client-tui/internal/tui/grpc"
)

type orderSuccessMsg struct{}

type orderErrorMsg struct {
	err error
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// 1️⃣ Mensagens globais (async)
	switch msg := msg.(type) {

	case orderSuccessMsg:
		m.Screen = MenuScreen
		m.FocusIndex = 0
		m.Err = nil

		m.CustomerID.SetValue("")
		m.ProductID.SetValue("")
		m.Quantity.SetValue("")
		m.UnitPrice.SetValue("")

		return m, nil

	case orderErrorMsg:
		m.Err = msg.err
		m.Screen = MenuScreen
		return m, nil
	}

	// 2️⃣ Lógica por tela
	switch m.Screen {

	case MenuScreen:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "up":
				if m.Cursor > 0 {
					m.Cursor--
				}
			case "down":
				if m.Cursor < len(m.Menu)-1 {
					m.Cursor++
				}
			case "enter":
				if m.Menu[m.Cursor] == "Novo Pedido" {
					m.Screen = FormScreen
				}
				if m.Menu[m.Cursor] == "Sair" {
					return m, tea.Quit
				}
			}
		}

	case FormScreen:
		inputs := []*textinput.Model{
			&m.CustomerID,
			&m.ProductID,
			&m.Quantity,
			&m.UnitPrice,
		}

		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {

			case "enter":
				if m.FocusIndex == len(inputs)-1 {
					return m, m.submitOrder()
				}
				m.FocusIndex++

			case "esc":
				m.Screen = MenuScreen
				m.FocusIndex = 0
			}
		}

		for i := range inputs {
			inputs[i].Blur()
			if i == m.FocusIndex {
				inputs[i].Focus()
			}

			var cmd tea.Cmd
			*inputs[i], cmd = inputs[i].Update(msg)
			_ = cmd
		}
	}

	return m, nil
}

func (m Model) submitOrder() tea.Cmd {
	return func() tea.Msg {

		customerID, _ := strconv.ParseInt(m.CustomerID.Value(), 10, 64)
		productID, _ := strconv.ParseInt(m.ProductID.Value(), 10, 64)
		qty, _ := strconv.Atoi(m.Quantity.Value())
		price, _ := strconv.ParseFloat(m.UnitPrice.Value(), 32)

		order := domain.Order{
			CustomerID: customerID,
			OrderItems: []domain.OrderItem{
				{
					ProductID: productID,
					Quantity:  int32(qty),
					UnitPrice: float32(price),
				},
			},
		}

		if err := grpc.PlaceOrder(order); err != nil {
			return orderErrorMsg{err}
		}

		return orderSuccessMsg{}
	}
}
