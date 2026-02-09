#!/bin/bash
set -e

PROTO_DIR="proto"
OUT_DIR="golang"

mkdir -p "$OUT_DIR"

protoc -I "$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go-grpc_out="$OUT_DIR" \
  "$PROTO_DIR"/*.proto

echo "Protos gerados em $OUT_DIR"
