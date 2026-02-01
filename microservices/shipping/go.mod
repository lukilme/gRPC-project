module ifpb.com/microservices/shipping

go 1.24.0

toolchain go1.24.12

require (
	github.com/go-sql-driver/mysql v1.9.3
	go.opentelemetry.io/otel/trace v1.39.0
	google.golang.org/grpc v1.78.0
	ifpb.com/microservices-proto/golang v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/sirupsen/logrus v1.9.4
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.39.0
	golang.org/x/sys v0.39.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace ifpb.com/microservices-proto/golang => ../../microservices-proto/golang
