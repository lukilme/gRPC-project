```sh
├── client
│   ├── domain
│   │   └── order.go
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   └── tui
│   │       ├── form.go
│   │       ├── grpc
│   │       │   └── client.go
│   │       ├── model.go
│   │       ├── update.go
│   │       └── view.go
│   └── main
│       └── main.go
├── curl
│   ├── order.proto
│   ├── order_test.proto
│   └── payment.proto
├── docker-compose.yml
├── init.sql
├── main.sh
├── Makefile
├── microservices
│   ├── order
│   │   ├── cmd
│   │   │   └── main.go
│   │   ├── config
│   │   │   └── config.go
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── internal
│   │   │   ├── adapters
│   │   │   │   ├── db
│   │   │   │   │   └── mysql.go
│   │   │   │   ├── grpc
│   │   │   │   │   └── server.go
│   │   │   │   └── payment
│   │   │   │       └── payment.go
│   │   │   └── application
│   │   │       ├── core
│   │   │       │   ├── api
│   │   │       │   │   └── api.go
│   │   │       │   └── domain
│   │   │       │       └── order.go
│   │   │       └── ports
│   │   │           └── payment.go
│   │   ├── proto
│   │   │   ├── order.proto
│   │   │   ├── payment.proto
│   │   │   └── shipping.proto
│   │   ├── proto copy
│   │   │   └── payment.proto
│   │   └── test
│   │       └── main_test.go
│   ├── payment
│   │   ├── cmd
│   │   │   └── main.go
│   │   ├── config
│   │   │   └── config.go
│   │   ├── DB_DRIVER=mysql
│   │   ├── deployment.yaml
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── internal
│   │       ├── adapters
│   │       │   ├── db
│   │       │   │   └── db.go
│   │       │   └── grpc
│   │       │       ├── grpc.go
│   │       │       └── server.go
│   │       ├── application
│   │       │   └── core
│   │       │       ├── api
│   │       │       │   └── api.go
│   │       │       └── domain
│   │       │           └── payment.go
│   │       └── ports
│   │           ├── api.go
│   │           └── db.go
│   └── shipping
│       ├── cmd
│       │   └── main.go
│       ├── config
│       │   └── config.go
│       ├── go.mod
│       ├── go.sum
│       └── internal
│           ├── adapters
│           │   ├── db
│           │   │   └── mysql.go
│           │   └── grpc
│           │       └── server.go
│           ├── application
│           │   ├── core
│           │   │   ├── api
│           │   │   │   └── api.go
│           │   │   └── domain
│           │   │       └── shipping.go
│           │   └── ports
│           ├── domain
│           └── ports
├── microservices-proto
│   ├── generator_proto.sh
│   ├── golang
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── order
│   │   │   ├── order_grpc.pb.go
│   │   │   └── order.pb.go
│   │   ├── payment
│   │   │   ├── payment_grpc.pb.go
│   │   │   └── payment.pb.go
│   │   └── shipping
│   │       ├── shipping_grpc.pb.go
│   │       └── shipping.pb.go
│   ├── order_grpc.pb.go
│   ├── order.pb.go
│   ├── payment_grpc.pb.go
│   ├── payment.pb.go
│   ├── proto
│   │   ├── order.proto
│   │   ├── payment.proto
│   │   └── shipping.proto
│   ├── shipping_grpc.pb.go
│   └── shipping.pb.go
├── README.md
├── run.sh
└── send.sh
```