package main

import (
	"log"

	"ifpb.com/microservices/order/internal/adapters/db"
)

func test() {
	adapter, err := db.NewAdapter("root:password@tcp(127.0.0.1:3306)/order_db?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer adapter.Close()

	exists, err := adapter.ItemExists(1)
	log.Printf("Item 1 exists: %v, error: %v", exists, err)

	exists, err = adapter.ItemExists(999)
	log.Printf("Item 999 exists: %v, error: %v", exists, err)

	items, err := adapter.ListItems()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Total items in stock: %d", len(items))
	for _, item := range items {
		log.Printf("  - %d: %s (R$ %.2f)", item.ID, item.Name, item.Price)
	}
}
