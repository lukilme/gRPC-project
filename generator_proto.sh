#!/bin/bash

PROTO_DIR="./microservices/order/proto"

OUT_DIR="./microservices-proto/golang/payment"

mkdir -p $OUT_DIR

protoc --go_out=$OUT_DIR --go-grpc_out=$OUT_DIR $PROTO_DIR/payment.proto