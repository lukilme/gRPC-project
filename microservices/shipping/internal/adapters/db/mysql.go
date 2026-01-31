package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"ifpb.com/microservices/shipping/internal/application/core/domain"
)

type Adapter struct {
	db *sql.DB
}

func NewAdapter(dataSourceURL string) (*Adapter, error) {
	db, err := sql.Open("mysql", dataSourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database connection established and tables created")
	return &Adapter{db: db}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{

		`CREATE TABLE IF NOT EXISTS shipping (
            id BIGINT AUTO_INCREMENT PRIMARY KEY,
            delivery_days INT NOT NULL,
            created_at BIGINT NOT NULL
        )`,

		`CREATE TABLE IF NOT EXISTS order_items (
			quantity INT NOT NULL,
			order_id BIGINT NOT NULLL
			CONSTRAINT fk_order_items_order
				FOREIGN KEY (order_id) REFERENCES orders(id)
				ON DELETE CASCADE
		);
`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
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

func (a *Adapter) Close() error {
	return a.db.Close()
}

func (a *Adapter) Save(shipping *domain.Shipping) error {
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

	shippingQuery := `INSERT INTO orders (customer_id, status, created_at) VALUES (?, ?, ?)`

	result, err := tx.Exec(shippingQuery, shipping.OrderID, shipping.DeliveryDays, time.Now().Unix())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert order: %w", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	shipping.OrderID = int(orderID)

	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)`

	for _, item := range shipping.Items {
		_, err := tx.Exec(itemQuery, orderID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Order saved successfully with id %d", orderID)
	return nil
}
