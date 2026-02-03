#  grpcurl -plaintext -d '{
#   "customer_id": 2,
#   "items": [
#     {"product_id": 100, "quantity": 2, "unit_price": 29.99},
#     {"product_id": 101, "quantity": 5, "unit_price": 59.99}
#   ]
# }' localhost:3000 payment.OrderService/PlaceOrder

 grpcurl -plaintext -d '{
  "customer_id": 2,
  "items": [
    {"product_id": 100, "quantity": 32, "unit_price": 13}
  ]
}' localhost:3000 payment.OrderService/PlaceOrder

 grpcurl -plaintext -d '{
  "order_id": "order-3",
  "items": [
    {"item_id": "item-001", "quantity": 30},
    {"item_id": "item-002", "quantity": 540}
  ]
}' localhost:3002 shipping.ShippingService/CalculateShipping