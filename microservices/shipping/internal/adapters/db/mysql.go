package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	_ "github.com/go-sql-driver/mysql"
	"ifpb.com/microservices/shipping/internal/application/core/domain"
)

// Colors for logs
const (
	greenColor  = "\033[32m"
	yellowColor = "\033[33m"
	redColor    = "\033[31m"
	blueColor   = "\033[34m"
	resetColor  = "\033[0m"
)

type Adapter struct {
	db *sql.DB
}

func (a *Adapter) GetByID(shippingID int) (*domain.Shipping, error) {
	panic("not implemented")
}

func (a *Adapter) Update(shipping *domain.Shipping) error {
	panic("not implemented")
}

func NewAdapter(dataSourceURL string) (*Adapter, error) {
	log.Printf("%sINFO: Initializing database adapter%s", blueColor, resetColor)
	log.Printf("%sDETAIL: Data source URL: %s%s", blueColor, maskPassword(dataSourceURL), resetColor)

	db, err := sql.Open("mysql", dataSourceURL)
	if err != nil {
		log.Printf("%sERROR: Failed to open database connection: %v%s", redColor, err, resetColor)

		// Check for specific connection errors
		if strings.Contains(err.Error(), "invalid DSN") {
			return nil, status.Error(codes.InvalidArgument,
				"Invalid database connection string. Please check the data source URL format")
		}
		if strings.Contains(err.Error(), "access denied") {
			return nil, status.Error(codes.PermissionDenied,
				"Database access denied. Please verify credentials")
		}

		return nil, status.Error(codes.Unavailable,
			"Database service is currently unavailable. Please try again later")
	}

	log.Printf("%sINFO: Testing database connection...%s", blueColor, resetColor)
	if err := db.Ping(); err != nil {
		log.Printf("%sERROR: Database connection test failed: %v%s", redColor, err, resetColor)

		if strings.Contains(err.Error(), "connection refused") {
			return nil, status.Error(codes.Unavailable,
				"Cannot connect to database server. Please ensure the database is running")
		}
		if strings.Contains(err.Error(), "unknown database") {
			return nil, status.Error(codes.FailedPrecondition,
				"Database does not exist. Please create the database first")
		}
		if strings.Contains(err.Error(), "timeout") {
			return nil, status.Error(codes.DeadlineExceeded,
				"Database connection timeout. Please check network connectivity")
		}

		return nil, status.Error(codes.Unavailable,
			"Unable to establish database connection. Please verify connection parameters")
	}

	log.Printf("%sINFO: Creating database tables...%s", blueColor, resetColor)
	if err := createTables(db); err != nil {
		log.Printf("%sERROR: Table creation failed: %v%s", redColor, err, resetColor)

		if strings.Contains(err.Error(), "syntax error") {
			return nil, status.Error(codes.FailedPrecondition,
				"Invalid table schema definition. Please check the table creation queries")
		}
		if strings.Contains(err.Error(), "privilege") {
			return nil, status.Error(codes.PermissionDenied,
				"Insufficient database privileges to create tables")
		}

		return nil, status.Error(codes.Internal,
			"Failed to initialize database schema. Please contact administrator")
	}

	log.Printf("%sSUCCESS: Database adapter initialized successfully%s", greenColor, resetColor)
	log.Printf("%sDETAIL: Connection pool stats: MaxOpenConns=%d, MaxIdleConns=%d%s",
		blueColor, db.Stats().MaxOpenConnections, db.Stats().Idle, resetColor)

	return &Adapter{db: db}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS shipping (
			id VARCHAR(50) PRIMARY KEY,
			order_id VARCHAR(50) NOT NULL UNIQUE,
			delivery_days INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS shipping_items (
			id VARCHAR(50) PRIMARY KEY,
			shipping_id VARCHAR(50) NOT NULL,
			item_id VARCHAR(50) NOT NULL,
			quantity INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (shipping_id) REFERENCES shipping(id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		tableName := extractTableName(query)
		log.Printf("%sINFO: Creating table '%s' if not exists...%s", blueColor, tableName, resetColor)

		_, err := db.Exec(query)
		if err != nil {
			// Check for specific table errors
			if strings.Contains(err.Error(), "already exists") {
				log.Printf("%sWARNING: Table '%s' already exists%s", yellowColor, tableName, resetColor)
				continue
			}
			if strings.Contains(err.Error(), "parse error") {
				log.Printf("%sERROR: SQL syntax error in table creation: %v%s", redColor, err, resetColor)
				return status.Errorf(codes.FailedPrecondition,
					"Invalid SQL syntax for table '%s': %v", tableName, err)
			}

			log.Printf("%sERROR: Failed to create table '%s': %v%s", redColor, tableName, err, resetColor)
			return status.Errorf(codes.Internal,
				"Failed to create database table '%s': %v", tableName, err)
		}

		log.Printf("%sSUCCESS: Table '%s' ready%s", greenColor, tableName, resetColor)
	}

	return nil
}

func (a *Adapter) Close() error {
	log.Printf("%sINFO: Closing database connection...%s", blueColor, resetColor)

	if a.db == nil {
		log.Printf("%sWARNING: Database connection is already nil%s", yellowColor, resetColor)
		return nil
	}

	stats := a.db.Stats()
	log.Printf("%sDETAIL: Connection stats before close: OpenConnections=%d, InUse=%d, Idle=%d%s",
		blueColor, stats.OpenConnections, stats.InUse, stats.Idle, resetColor)

	if err := a.db.Close(); err != nil {
		log.Printf("%sERROR: Failed to close database connection gracefully: %v%s", redColor, err, resetColor)

		if strings.Contains(err.Error(), "bad connection") {
			return status.Error(codes.FailedPrecondition,
				"Database connection was already in a bad state")
		}

		return status.Error(codes.Internal,
			"Failed to properly close database connection. Some resources may not be released")
	}

	log.Printf("%sSUCCESS: Database connection closed successfully%s", greenColor, resetColor)
	return nil
}

func (a *Adapter) Save(shipping *domain.Shipping) error {
	log.Printf("%sINFO: Saving shipping record: OrderID=%s, Items=%d%s",
		blueColor, shipping.OrderID, len(shipping.Items), resetColor)

	if shipping.OrderID == "" {
		log.Printf("%sERROR: Cannot save shipping - OrderID is empty%s", redColor, resetColor)
		return status.Error(codes.InvalidArgument, "Order ID is required")
	}
	if len(shipping.Items) == 0 {
		log.Printf("%sERROR: Cannot save shipping - no items provided for order %s%s",
			redColor, shipping.OrderID, resetColor)
		return status.Error(codes.InvalidArgument, "At least one item is required")
	}

	tx, err := a.db.Begin()
	if err != nil {
		log.Printf("%sERROR: Failed to begin database transaction: %v%s", redColor, err, resetColor)

		if strings.Contains(err.Error(), "connection refused") {
			return status.Error(codes.Unavailable,
				"Database connection lost. Please try again")
		}
		if strings.Contains(err.Error(), "too many connections") {
			return status.Error(codes.ResourceExhausted,
				"Database connection pool exhausted. Please try again later")
		}

		return status.Error(codes.Internal,
			"Failed to start database transaction")
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			log.Printf("%sWARNING: Database transaction panic recovered: %v%s", yellowColor, p, resetColor)
		}
	}()

	if shipping.ID == "" {
		shipping.ID = generateShippingID()
		log.Printf("%sINFO: Generated Shipping ID: %s%s", greenColor, shipping.ID, resetColor)
	}

	shippingQuery := `INSERT INTO shipping (id, order_id, delivery_days, created_at) VALUES (?, ?, ?, ?)`
	log.Printf("%sDETAIL: Executing shipping insert: %s%s", blueColor, shippingQuery, resetColor)

	_, err = tx.Exec(shippingQuery, shipping.ID, shipping.OrderID, shipping.DeliveryDays, time.Now())
	if err != nil {
		tx.Rollback()

		if strings.Contains(err.Error(), "Duplicate entry") {
			log.Printf("%sERROR: Duplicate shipping entry for order %s%s", redColor, shipping.OrderID, resetColor)
			return status.Errorf(codes.AlreadyExists,
				"Shipping record already exists for order '%s'", shipping.OrderID)
		}
		if strings.Contains(err.Error(), "foreign key constraint") {
			log.Printf("%sERROR: Foreign key constraint violation for shipping %s%s", redColor, shipping.ID, resetColor)
			return status.Error(codes.FailedPrecondition,
				"Database constraint violation")
		}

		log.Printf("%sERROR: Failed to insert shipping record: %v%s", redColor, err, resetColor)
		return status.Error(codes.Internal,
			"Failed to save shipping information")
	}

	itemQuery := `INSERT INTO shipping_items (id, shipping_id, item_id, quantity, created_at) VALUES (?, ?, ?, ?, ?)`
	log.Printf("%sINFO: Inserting %d items for shipping %s...%s",
		blueColor, len(shipping.Items), shipping.ID, resetColor)

	for i, item := range shipping.Items {
		if item.Quantity <= 0 {
			tx.Rollback()
			log.Printf("%sERROR: Invalid quantity for item %s: %d%s",
				redColor, item.ItemID, item.Quantity, resetColor)
			return status.Errorf(codes.InvalidArgument,
				"Invalid quantity for item '%s': must be greater than zero", item.ItemID)
		}

		itemID := generateShippingItemID()
		log.Printf("%sDETAIL: Item %d: %s (quantity: %d)%s",
			blueColor, i+1, item.ItemID, item.Quantity, resetColor)

		_, err := tx.Exec(itemQuery, itemID, shipping.ID, item.ItemID, item.Quantity, time.Now())
		if err != nil {
			tx.Rollback()

			if strings.Contains(err.Error(), "Duplicate entry") {
				log.Printf("%sERROR: Duplicate item entry: %s%s", redColor, item.ItemID, resetColor)
				return status.Errorf(codes.AlreadyExists,
					"Item '%s' already exists in this shipping", item.ItemID)
			}

			log.Printf("%sERROR: Failed to insert shipping item %s: %v%s", redColor, item.ItemID, err, resetColor)
			return status.Errorf(codes.Internal,
				"Failed to save item '%s'", item.ItemID)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("%sERROR: Failed to commit transaction: %v%s", redColor, err, resetColor)

		if strings.Contains(err.Error(), "transaction has already been committed or rolled back") {
			return status.Error(codes.FailedPrecondition,
				"Transaction was already completed")
		}

		return status.Error(codes.Internal,
			"Failed to finalize shipping save operation")
	}

	log.Printf("%sSUCCESS: Shipping saved: ID=%s, OrderID=%s, Items=%d, DeliveryDays=%d%s",
		greenColor, shipping.ID, shipping.OrderID, len(shipping.Items), shipping.DeliveryDays, resetColor)
	return nil
}

func (a *Adapter) Get(orderID int) (*domain.Shipping, error) {
	log.Printf("%sINFO: Retrieving shipping for order ID: %d%s", blueColor, orderID, resetColor)

	if orderID <= 0 {
		log.Printf("%sERROR: Invalid order ID: %d%s", redColor, orderID, resetColor)
		return nil, status.Error(codes.InvalidArgument,
			fmt.Sprintf("Invalid order ID: %d. Must be a positive number", orderID))
	}

	shippingQuery := `SELECT id, order_id, delivery_days, created_at FROM shipping WHERE order_id = ?`
	log.Printf("%sDETAIL: Executing query: %s with param: %d%s", blueColor, shippingQuery, orderID, resetColor)

	var shipping domain.Shipping
	err := a.db.QueryRow(shippingQuery, orderID).Scan(
		&shipping.ID,
		&shipping.OrderID,
		&shipping.DeliveryDays,
		&shipping.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("%sWARNING: No shipping found for order ID: %d%s", yellowColor, orderID, resetColor)
			return nil, nil
		}

		log.Printf("%sERROR: Database query failed for order %d: %v%s", redColor, orderID, err, resetColor)

		if strings.Contains(err.Error(), "connection") {
			return nil, status.Error(codes.Unavailable,
				"Database connection error while retrieving shipping")
		}

		return nil, status.Error(codes.Internal,
			"Failed to retrieve shipping information")
	}

	itemsQuery := `SELECT item_id, quantity FROM shipping_items WHERE shipping_id = ? ORDER BY created_at`
	log.Printf("%sDETAIL: Retrieving items for shipping %s...%s", blueColor, shipping.ID, resetColor)

	rows, err := a.db.Query(itemsQuery, shipping.ID)
	if err != nil {
		log.Printf("%sERROR: Failed to retrieve shipping items: %v%s", redColor, err, resetColor)
		return nil, status.Error(codes.Internal,
			"Failed to retrieve shipping items")
	}
	defer rows.Close()

	var items []domain.OrderItem
	itemCount := 0
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ItemID, &item.Quantity); err != nil {
			log.Printf("%sERROR: Failed to scan item row: %v%s", redColor, err, resetColor)
			return nil, status.Error(codes.Internal,
				"Failed to process shipping item data")
		}
		items = append(items, item)
		itemCount++
	}

	if err = rows.Err(); err != nil {
		log.Printf("%sERROR: Error iterating through item rows: %v%s", redColor, err, resetColor)
		return nil, status.Error(codes.Internal,
			"Error processing shipping items")
	}

	shipping.Items = items
	log.Printf("%sSUCCESS: Shipping retrieved: ID=%s, OrderID=%s, Items=%d, DeliveryDays=%d%s",
		greenColor, shipping.ID, shipping.OrderID, len(items), shipping.DeliveryDays, resetColor)

	return &shipping, nil
}

func (a *Adapter) CalculateDeliveryDays(items []domain.OrderItem) (int, error) {
	log.Printf("%sINFO: Calculating delivery days for %d items%s", blueColor, len(items), resetColor)

	if len(items) == 0 {
		log.Printf("%sWARNING: No items provided for delivery calculation%s", yellowColor, resetColor)
		return 1, nil
	}

	totalUnits := 0
	for i, item := range items {
		if item.Quantity <= 0 {
			log.Printf("%sERROR: Invalid quantity for item %s: %d%s",
				redColor, item.ItemID, item.Quantity, resetColor)
			return 0, status.Errorf(codes.InvalidArgument,
				"Invalid quantity for item '%s': must be greater than zero", item.ItemID)
		}
		totalUnits += item.Quantity
		log.Printf("%sDETAIL: Item %d: %s (quantity: %d, running total: %d)%s",
			blueColor, i+1, item.ItemID, item.Quantity, totalUnits, resetColor)
	}

	deliveryDays := 1 + (totalUnits / 5)
	if deliveryDays < 1 {
		deliveryDays = 1
	}

	log.Printf("%sSUCCESS: Delivery calculation complete: TotalUnits=%d, DeliveryDays=%d%s",
		greenColor, totalUnits, deliveryDays, resetColor)
	return deliveryDays, nil
}

func generateShippingID() string {
	return fmt.Sprintf("shp-%d", time.Now().UnixNano())
}

func generateShippingItemID() string {
	return fmt.Sprintf("shp-item-%d", time.Now().UnixNano()+rand.Int63n(1000))
}

func maskPassword(dataSourceURL string) string {
	if strings.Contains(dataSourceURL, "@") {
		parts := strings.Split(dataSourceURL, "@")
		if len(parts) > 1 {
			credentials := strings.Split(parts[0], ":")
			if len(credentials) > 1 {
				return strings.Replace(dataSourceURL, credentials[1], "****", 1)
			}
		}
	}
	return dataSourceURL
}

func extractTableName(query string) string {
	query = strings.ToUpper(query)
	if strings.Contains(query, "CREATE TABLE") {
		parts := strings.Fields(query)
		for i, part := range parts {
			if part == "TABLE" && i+1 < len(parts) {
				tableName := strings.Trim(parts[i+1], "`")
				return strings.Split(tableName, "(")[0]
			}
		}
	}
	return "unknown"
}
