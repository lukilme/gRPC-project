#!/bin/bash

PROTO_DIR="./proto"
OUT_DIR="./microservices-proto/golang"

mkdir -p "$OUT_DIR"

protoc -I"$PROTO_DIR" \
  --go_out=paths=source_relative:"$OUT_DIR" \
  --go-grpc_out=paths=source_relative:"$OUT_DIR" \
  "$PROTO_DIR"/order.proto

protoc -I"$PROTO_DIR" \
  --go_out=paths=source_relative:"$OUT_DIR" \
  --go-grpc_out=paths=source_relative:"$OUT_DIR" \
  "$PROTO_DIR"/payment.proto

protoc -I"$PROTO_DIR" \
  --go_out=paths=source_relative:"$OUT_DIR" \
  --go-grpc_out=paths=source_relative:"$OUT_DIR" \
  "$PROTO_DIR"/shipping.proto
echo "Protos gerados em $OUT_DIR"