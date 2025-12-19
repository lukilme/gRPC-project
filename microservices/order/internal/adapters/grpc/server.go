package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"ifpb.com/microservices/order/internal/application/core/api"
)

type Adapter struct {
	api    *api.Application
	port   string
	server *grpc.Server
}

func NewAdapter(api *api.Application, port string) *Adapter {
	return &Adapter{
		api:  api,
		port: port,
	}
}

func (a *Adapter) Run() {
	lis, err := net.Listen("tcp", ":"+a.port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	a.server = grpc.NewServer()
	reflection.Register(a.server)
	log.Printf("gRPC server listening on port %s", a.port)
	if err := a.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (a *Adapter) Stop() {
	if a.server != nil {
		a.server.GracefulStop()
	}
}
