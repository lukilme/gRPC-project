package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"ifpb.com/microservices/order/internal/application/core/domain"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
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

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Printf("%sDatabase connection established and tables created%s", colorGreen, colorReset)
	return &Adapter{db: db}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`DROP TABLE IF EXISTS order_items`,
		`DROP TABLE IF EXISTS items`,
		`DROP TABLE IF EXISTS orders`,

		`CREATE TABLE orders (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			customer_id BIGINT NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			total_amount DECIMAL(10, 2) DEFAULT 0.0,
			delivery_days INT DEFAULT 0,
			created_at BIGINT NOT NULL
		)`,

		`CREATE TABLE order_items (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			order_id BIGINT NOT NULL,
			product_id BIGINT NOT NULL,
			quantity INT NOT NULL,
			unit_price DECIMAL(10, 2) NOT NULL,
			CONSTRAINT fk_order_items_order
				FOREIGN KEY (order_id) REFERENCES orders(id)
				ON DELETE CASCADE
		)`,

		`CREATE TABLE items (
			id BIGINT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10, 2) NOT NULL,
			created_at BIGINT NOT NULL
		)`,
	}

	for _, query := range queries {
		log.Printf("%sExecuting query: %s%s", colorBlue, query, colorReset)
		_, err := db.Exec(query)
		if err != nil {
			if !contains(err.Error(), "already exists") && !contains(err.Error(), "Duplicate key") {
				return fmt.Errorf("failed to execute query: %s, error: %w", query, err)
			}
			log.Printf("%sTable warning: %v%s", colorYellow, err, colorReset)
		}
	}

	seedItems(db)

	return nil
}

func seedItems(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM items").Scan(&count)

	if count == 0 {
		log.Printf("%sPopulating items table with initial data...%s", colorBlue, colorReset)

		items := []struct {
			id    int64
			name  string
			price float32
		}{
			{1, "Notebook Dell Inspiron", 4500.00},
			{2, "Mouse Logitech MX Master", 150.00},
			{3, "Mechanical Keyboard Redragon", 350.00},
			{4, "Monitor LG 24\" Full HD", 1200.00},
			{5, "Webcam Logitech C920", 250.00},
			{6, "Headset HyperX Cloud II", 400.00},
			{7, "SSD Kingston 500GB", 300.00},
			{8, "RAM Memory 16GB DDR4", 350.00},
		}

		for _, item := range items {
			_, err := db.Exec(
				"INSERT INTO items (id, name, price, created_at) VALUES (?, ?, ?, ?)",
				item.id, item.name, item.price, time.Now().Unix(),
			)
			if err != nil {
				log.Printf("%sError inserting item %d: %v%s", colorYellow, item.id, err, colorReset)
			} else {
				log.Printf("%s  Item %d inserted: %s%s", colorGreen, item.id, item.name, colorReset)
			}
		}
		log.Printf("%sItems table populated successfully%s", colorGreen, colorReset)
	} else {
		log.Printf("%sItems table already contains %d records%s", colorBlue, count, colorReset)
	}
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
	log.Printf("%sDatabase connection closed%s", colorBlue, colorReset)
	return nil
}

func (a *Adapter) Get(id int64) (domain.Order, error) {
	log.Printf("%sSearching order with ID: %d%s", colorBlue, id, colorReset)

	orderQuery := `SELECT id, customer_id, status, total_amount, delivery_days, created_at FROM orders WHERE id = ?`

	row := a.db.QueryRow(orderQuery, id)

	var order domain.Order
	err := row.Scan(&order.ID, &order.CustomerID, &order.Status, &order.TotalAmount, &order.DeliveryDays, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Order{}, fmt.Errorf("order not found with id %d", id)
		}
		return domain.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

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
	log.Printf("%sOrder %d found with %d items%s", colorGreen, order.ID, len(order.OrderItems), colorReset)
	return order, nil
}

func (a *Adapter) Save(order *domain.Order) error {
	log.Printf("%sSaving order for customer: %d%s", colorYellow, order.CustomerID, colorReset)

	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Printf("%sPanic recovered in Save: %v%s", colorRed, p, colorReset)
		}
	}()

	if order.TotalAmount == 0 {
		order.TotalAmount = order.TotalPrice()
	}

	if order.Status == "" {
		order.Status = "pending"
	}

	if order.CreatedAt == 0 {
		order.CreatedAt = time.Now().Unix()
	}

	orderQuery := `INSERT INTO orders (customer_id, status, total_amount, delivery_days, created_at) VALUES (?, ?, ?, ?, ?)`

	result, err := tx.Exec(orderQuery, order.CustomerID, order.Status, order.TotalAmount, order.DeliveryDays, order.CreatedAt)
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

	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)`

	for i, item := range order.OrderItems {
		_, err := tx.Exec(itemQuery, orderID, item.ProductID, item.Quantity, item.UnitPrice)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert order item %d (product_id: %d): %w", i+1, item.ProductID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("%sOrder saved successfully: ID=%d, Total=%.2f, Items=%d%s",
		colorGreen, order.ID, order.TotalAmount, len(order.OrderItems), colorReset)
	return nil
}

func (a *Adapter) Update(order *domain.Order) error {
	log.Printf("%sUpdating order %d%s", colorYellow, order.ID, colorReset)

	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Printf("%sPanic recovered in Update: %v%s", colorRed, p, colorReset)
		}
	}()

	updateOrderQuery := `UPDATE orders SET status = ?, total_amount = ?, delivery_days = ? WHERE id = ?`

	result, err := tx.Exec(updateOrderQuery, order.Status, order.TotalAmount, order.DeliveryDays, order.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("order with id %d not found", order.ID)
	}

	deleteItemsQuery := `DELETE FROM order_items WHERE order_id = ?`
	_, err = tx.Exec(deleteItemsQuery, order.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete old order items: %w", err)
	}

	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES (?, ?, ?, ?)`

	for i, item := range order.OrderItems {
		_, err := tx.Exec(itemQuery, order.ID, item.ProductID, item.Quantity, item.UnitPrice)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert order item %d: %w", i+1, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("%sOrder %d updated successfully. Status: %s%s", colorGreen, order.ID, order.Status, colorReset)
	return nil
}

func (a *Adapter) ItemExists(productID int64) (bool, error) {
	log.Printf("%sChecking if item %d exists in stock%s", colorBlue, productID, colorReset)

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM items WHERE id = ?)`

	err := a.db.QueryRow(query, productID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check item existence: %w", err)
	}

	if exists {
		log.Printf("%sItem %d exists in stock%s", colorGreen, productID, colorReset)
	} else {
		log.Printf("%sItem %d does NOT exist in stock%s", colorRed, productID, colorReset)
	}

	return exists, nil
}

func (a *Adapter) ListItems() ([]struct {
	ID    int64
	Name  string
	Price float32
}, error) {
	rows, err := a.db.Query("SELECT id, name, price FROM items ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []struct {
		ID    int64
		Name  string
		Price float32
	}

	for rows.Next() {
		var item struct {
			ID    int64
			Name  string
			Price float32
		}
		if err := rows.Scan(&item.ID, &item.Name, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
