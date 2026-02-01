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
			order_id BIGINT AUTO_INCREMENT NOT NULL UNIQUE,
			delivery_days INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_order_id (order_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS shipping_items (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			shipping_id BIGINT AUTO_INCREMENT NOT NULL,
			item_id BIGINT AUTO_INCREMENT NOT NULL,
			quantity INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT fk_shipping_items_shipping
				FOREIGN KEY (shipping_id) REFERENCES shipping(id)
				ON DELETE CASCADE,
			INDEX idx_shipping_id (shipping_id),
			INDEX idx_item_id (item_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			if !isTableExistsError(err) {
				return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
			}
			log.Printf("Table already exists or warning: %v", err)
		}
	}
	return nil
}

func isTableExistsError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "already exists") ||
		contains(errStr, "Duplicate") ||
		contains(errStr, "exists")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (a *Adapter) Close() error {
	if err := a.db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
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

	shippingQuery := `INSERT INTO shipping (id, order_id, delivery_days, created_at) VALUES (?, ?, ?, ?)`

	result, err := tx.Exec(shippingQuery,
		shipping.OrderID,
		shipping.DeliveryDays,
		time.Now(),
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert shipping: %w", err)
	}

	itemQuery := `INSERT INTO shipping_items (id, shipping_id, item_id, quantity, created_at) VALUES (?, ?, ?, ?, ?)`
	shippingID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	for _, item := range shipping.Items {

		_, err := tx.Exec(itemQuery,
			shippingID,
			item.ItemID,
			item.Quantity,
			time.Now(),
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert shipping item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Shipping saved successfully for order %s", shipping.OrderID)
	return nil
}

func (a *Adapter) Get(orderID int) (*domain.Shipping, error) {
	shippingQuery := `SELECT id, order_id, delivery_days, created_at FROM shipping WHERE order_id = ?`

	var shipping domain.Shipping
	err := a.db.QueryRow(shippingQuery, orderID).Scan(
		&shipping.ID,
		&shipping.OrderID,
		&shipping.DeliveryDays,
		&shipping.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shipping: %w", err)
	}

	itemsQuery := `SELECT item_id, quantity FROM shipping_items WHERE shipping_id = ? ORDER BY created_at`

	rows, err := a.db.Query(itemsQuery, shipping.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shipping items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ItemID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	shipping.Items = items
	return &shipping, nil
}

func (a *Adapter) CalculateDeliveryDays(items []domain.OrderItem) (int, error) {
	totalUnits := 0
	for _, item := range items {
		totalUnits += item.Quantity
	}

	deliveryDays := 1 + (totalUnits / 5)
	if deliveryDays < 1 {
		deliveryDays = 1
	}

	return deliveryDays, nil
}
