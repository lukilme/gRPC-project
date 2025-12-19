package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"ifpb.com/microservices/order/internal/application/core/domain"
)

type Adapter struct {
	db *sql.DB
}

func NewAdapter(dataSourceURL string) (*Adapter, error) {
	db, err := sql.Open("mysql", dataSourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Testar conexão
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Criar tabelas separadamente
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database connection established and tables created")
	return &Adapter{db: db}, nil
}

func createTables(db *sql.DB) error {
	// Executar cada comando SQL SEPARADAMENTE
	queries := []string{
		// Tabela orders
		`CREATE TABLE IF NOT EXISTS orders (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            customer_id BIGINT NOT NULL,
            status VARCHAR(50) NOT NULL DEFAULT 'pending',
            created_at BIGINT NOT NULL
        )`,

		// Tabela order_items
		`CREATE TABLE IF NOT EXISTS order_items (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            order_id BIGINT NOT NULL,
            product_id BIGINT NOT NULL,
            quantity INT NOT NULL,
            unit_price FLOAT NOT NULL
        )`,

		// Adicionar foreign key separadamente (opcional)
		`CREATE TABLE IF NOT EXISTS order_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    unit_price FLOAT NOT NULL,
    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id) REFERENCES orders(id)
        ON DELETE CASCADE
);
`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			// Ignorar erro se foreign key já existir
			if !contains(err.Error(), "already exists") && !contains(err.Error(), "Duplicate key") {
				return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
			}
			log.Printf("Warning: %v", err)
		}
	}
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func (a *Adapter) Get(id int64) (domain.Order, error) {
	// Primeiro, buscar a ordem
	orderQuery := `SELECT id, customer_id, status, created_at FROM orders WHERE id = ?`

	row := a.db.QueryRow(orderQuery, id)

	var order domain.Order
	err := row.Scan(&order.ID, &order.CustomerID, &order.Status, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Order{}, fmt.Errorf("order not found with id %d", id)
		}
		return domain.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	// Buscar itens do pedido
	itemsQuery := `SELECT product_id, quantity, unit_price FROM order_items WHERE order_id = ?`

	rows, err := a.db.Query(itemsQuery, id)
	if err != nil {
		return order, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ProductID, &item.Quantity, &item.UnitPrice); err != nil {
			return order, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return order, fmt.Errorf("error iterating order items: %w", err)
	}

	order.OrderItems = items
	return order, nil
}

func (a *Adapter) Save(order *domain.Order) error {
	// Usar transação
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// Inserir ordem
	orderQuery := `INSERT INTO orders (customer_id, status, created_at) VALUES (?, ?, ?)`

	result, err := tx.Exec(orderQuery, order.CustomerID, order.Status, time.Now().Unix())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert order: %w", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	order.ID = orderID

	// Inserir itens
	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)`

	for _, item := range order.OrderItems {
		_, err := tx.Exec(itemQuery, orderID, item.ProductID, item.Quantity, item.UnitPrice)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Order saved successfully with id %d", orderID)
	return nil
}

// Método para fechar conexão
func (a *Adapter) Close() error {
	return a.db.Close()
}
