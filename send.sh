 grpcurl -plaintext -d '{
  "customer_id": 2,
  "items": [
    {"product_id": 100, "quantity": 2, "unit_price": 29.99},
    {"product_id": 101, "quantity": 5, "unit_price": 59.99}
  ]
}' localhost:3000 payment.OrderService/PlaceOrder