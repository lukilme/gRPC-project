package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	shippingpb "ifpb.com/microservices-proto/golang/shipping"
	"ifpb.com/microservices/shipping/internal/application/core/api"
	"ifpb.com/microservices/shipping/internal/application/core/domain"
)

// Colors for logs
const (
	blueColor   = "\033[34m"
	greenColor  = "\033[32m"
	yellowColor = "\033[33m"
	redColor    = "\033[31m"
	resetColor  = "\033[0m"
)

type Adapter struct {
	shippingpb.UnimplementedShippingServiceServer
	api    *api.Application
	port   int
	server *grpc.Server
}

func NewAdapter(api *api.Application, port int) *Adapter {
	log.Printf("%sINFO: Initializing gRPC Adapter on port %d%s", blueColor, port, resetColor)
	return &Adapter{
		api:  api,
		port: port,
	}
}

func (a *Adapter) CalculateShipping(ctx context.Context, req *shippingpb.ShippingRequest) (*shippingpb.ShippingResponse, error) {
	log.Printf("%sINFO: Received shipping calculation request for order: %s%s",
		blueColor, req.GetOrderId(), resetColor)
	log.Printf("%sDETAIL: Processing %d items for order %s%s",
		blueColor, len(req.GetItems()), req.GetOrderId(), resetColor)

	if req.GetOrderId() == "" {
		log.Printf("%sERROR: Invalid request - order_id is required but was empty%s", redColor, resetColor)
		return nil, status.Error(codes.InvalidArgument, "Order ID is required and cannot be empty")
	}

	if len(req.GetItems()) == 0 {
		log.Printf("%sERROR: Invalid request - order %s has no items%s",
			redColor, req.GetOrderId(), resetColor)
		return nil, status.Error(codes.InvalidArgument, "At least one item is required for shipping calculation")
	}

	domainItems := make([]domain.OrderItem, len(req.GetItems()))
	log.Printf("%sDETAIL: Converting protobuf items to domain items...%s", blueColor, resetColor)

	for i, pbItem := range req.GetItems() {
		if pbItem.GetQuantity() <= 0 {
			log.Printf("%sERROR: Invalid quantity for item %s: %d (must be positive)%s",
				redColor, pbItem.GetItemId(), pbItem.GetQuantity(), resetColor)
			return nil, status.Errorf(codes.InvalidArgument,
				"Invalid quantity for item '%s': %d. Quantity must be greater than zero",
				pbItem.GetItemId(), pbItem.GetQuantity())
		}

		if pbItem.GetItemId() == "" {
			log.Printf("%sERROR: Item at position %d has empty item_id%s", redColor, i+1, resetColor)
			return nil, status.Errorf(codes.InvalidArgument,
				"Item at position %d has empty item ID. All items must have a valid identifier", i+1)
		}

		domainItems[i] = domain.OrderItem{
			ItemID:   pbItem.GetItemId(),
			Quantity: int(pbItem.GetQuantity()),
		}
		log.Printf("%sDETAIL: Item %d: %s (quantity: %d)%s",
			blueColor, i+1, pbItem.GetItemId(), pbItem.GetQuantity(), resetColor)
	}

	log.Printf("%sDETAIL: Calling application layer for shipping calculation...%s", blueColor, resetColor)
	shipping, err := a.api.CalculateAndPlaceShipping(req.OrderId, domainItems)

	if err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("%sERROR: Application layer returned gRPC status: %v - %s%s",
				redColor, st.Code(), st.Message(), resetColor)

			switch st.Code() {
			case codes.AlreadyExists:
				log.Printf("%sDETAIL: Shipping already exists for order %s%s",
					blueColor, req.GetOrderId(), resetColor)
				return nil, status.Error(codes.AlreadyExists, st.Message())
			case codes.InvalidArgument:
				log.Printf("%sDETAIL: Validation error from application layer%s", blueColor, resetColor)
				return nil, status.Error(codes.InvalidArgument, st.Message())
			case codes.ResourceExhausted:
				log.Printf("%sERROR: System resources exhausted for order %s%s",
					redColor, req.GetOrderId(), resetColor)
				return nil, status.Error(codes.ResourceExhausted, st.Message())
			case codes.Unavailable:
				log.Printf("%sERROR: Service unavailable for order %s%s",
					redColor, req.GetOrderId(), resetColor)
				return nil, status.Error(codes.Unavailable, st.Message())
			case codes.Internal:
				log.Printf("%sERROR: Internal server error for order %s%s",
					redColor, req.GetOrderId(), resetColor)
				return nil, status.Error(codes.Internal,
					"An internal error occurred while processing your request. Please try again.")
			default:
				log.Printf("%sERROR: Unhandled gRPC code %v for order %s: %s%s",
					redColor, st.Code(), req.GetOrderId(), st.Message(), resetColor)
				return nil, st.Err()
			}
		}

		log.Printf("%sERROR: Non-gRPC error for order %s: %v%s",
			redColor, req.GetOrderId(), err, resetColor)

		if err.Error() == "must have at least one item" {
			return nil, status.Error(codes.InvalidArgument, "Order must contain at least one item")
		}
		if err.Error() == "quantity must be greater than zero" {
			return nil, status.Error(codes.InvalidArgument, "All item quantities must be greater than zero")
		}

		return nil, status.Error(codes.Internal,
			"An unexpected error occurred while calculating shipping")
	}

	log.Printf("%sSUCCESS: Shipping calculated for order %s: %d delivery days%s",
		greenColor, shipping.OrderID, shipping.DeliveryDays, resetColor)
	log.Printf("%sDETAIL: Shipping ID: %s, Items processed: %d%s",
		blueColor, shipping.ID, len(domainItems), resetColor)

	return &shippingpb.ShippingResponse{
		OrderId:      req.OrderId,
		DeliveryDays: int32(shipping.DeliveryDays),
		Status:       "CALCULATED",
	}, nil
}

func (a *Adapter) Run() {
	log.Printf("%sINFO: Starting gRPC server initialization%s", blueColor, resetColor)
	address := fmt.Sprintf(":%d", a.port)

	log.Printf("%sDETAIL: Attempting to listen on TCP address: %s%s", blueColor, address, resetColor)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("%sFATAL: Failed to listen on port %d: %v%s", redColor, a.port, err, resetColor)
		// Check for specific network errors
		if err.Error() == "bind: address already in use" {
			log.Printf("%sFATAL: Port %d is already in use%s", redColor, a.port, resetColor)
			log.Fatalf("Port %d is unavailable. Please check if another service is running on this port.", a.port)
		}
		log.Fatalf("Failed to start gRPC server on port %d: %v", a.port, err)
	}

	log.Printf("%sDETAIL: Creating gRPC server with logging interceptor%s", blueColor, resetColor)
	a.server = grpc.NewServer(
		grpc.UnaryInterceptor(a.loggingInterceptor()),
	)

	log.Printf("%sDETAIL: Registering Shipping Service server%s", blueColor, resetColor)
	shippingpb.RegisterShippingServiceServer(a.server, a)

	log.Printf("%sDETAIL: Registering gRPC reflection service%s", blueColor, resetColor)
	reflection.Register(a.server)

	log.Printf("%sSUCCESS: Shipping Service gRPC server listening on port %d%s",
		greenColor, a.port, resetColor)
	log.Printf("%sDETAIL: Server ready to accept connections%s", blueColor, resetColor)
	log.Printf("%sINFO: Server address: localhost:%d%s", blueColor, a.port, resetColor)

	log.Printf("%sINFO: Starting gRPC server serve loop%s", blueColor, resetColor)
	if err := a.server.Serve(lis); err != nil {
		log.Printf("%sFATAL: gRPC server failed: %v%s", redColor, err, resetColor)

		// Handle specific serve errors
		if err.Error() == "grpc: the server has been stopped" {
			log.Printf("%sINFO: Server stopped gracefully%s", greenColor, resetColor)
		} else {
			log.Printf("%sFATAL: Unexpected server failure: %v%s", redColor, err, resetColor)
			log.Fatalf("gRPC server encountered a fatal error: %v", err)
		}
	}
}

func (a *Adapter) Stop() {
	log.Printf("%sINFO: Stopping Shipping Service gRPC server...%s", yellowColor, resetColor)

	if a.server == nil {
		log.Printf("%sWARNING: Server instance is nil, nothing to stop%s", yellowColor, resetColor)
		return
	}

	log.Printf("%sDETAIL: Initiating graceful shutdown%s", blueColor, resetColor)
	a.server.GracefulStop()
	log.Printf("%sSUCCESS: gRPC server stopped successfully%s", greenColor, resetColor)
	log.Printf("%sDETAIL: All pending requests completed%s", blueColor, resetColor)
}

func (a *Adapter) loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod
		log.Printf("%sINFO: gRPC Request received: %s%s", blueColor, method, resetColor)

		if method == "/shipping.ShippingService/CalculateShipping" {
			if shippingReq, ok := req.(*shippingpb.ShippingRequest); ok {
				log.Printf("%sDETAIL: Request details - Order: %s, Items: %d%s",
					blueColor, shippingReq.GetOrderId(), len(shippingReq.GetItems()), resetColor)

				for i, item := range shippingReq.GetItems() {
					log.Printf("%sDETAIL:   Item %d: %s (quantity: %d)%s",
						blueColor, i+1, item.GetItemId(), item.GetQuantity(), resetColor)
				}
			}
		}

		log.Printf("%sDETAIL: Processing gRPC handler...%s", blueColor, resetColor)
		resp, err := handler(ctx, req)

		if err != nil {
			// Extract error code for better logging
			if st, ok := status.FromError(err); ok {
				switch st.Code() {
				case codes.InvalidArgument:
					log.Printf("%sWARNING: Invalid argument in %s: %v%s",
						yellowColor, method, st.Message(), resetColor)
				case codes.ResourceExhausted:
					log.Printf("%sERROR: Resource exhausted in %s: %v%s",
						redColor, method, st.Message(), resetColor)
				case codes.Unavailable:
					log.Printf("%sERROR: Service unavailable in %s: %v%s",
						redColor, method, st.Message(), resetColor)
				case codes.DeadlineExceeded:
					log.Printf("%sWARNING: Deadline exceeded in %s: %v%s",
						yellowColor, method, st.Message(), resetColor)
				case codes.Internal:
					log.Printf("%sERROR: Internal server error in %s: %v%s",
						redColor, method, st.Message(), resetColor)
				default:
					log.Printf("%sERROR: gRPC Error in %s (code: %v): %v%s",
						redColor, method, st.Code(), st.Message(), resetColor)
				}
			} else {
				log.Printf("%sERROR: Non-gRPC error in %s: %v%s",
					redColor, method, err, resetColor)
			}
		} else {
			log.Printf("%sSUCCESS: gRPC Request completed successfully: %s%s",
				greenColor, method, resetColor)

			if method == "/shipping.ShippingService/CalculateShipping" {
				if shippingResp, ok := resp.(*shippingpb.ShippingResponse); ok {
					log.Printf("%sDETAIL: Response details - Delivery Days: %d, Status: %s%s",
						blueColor, shippingResp.GetDeliveryDays(), shippingResp.GetStatus(), resetColor)
				}
			}
		}

		return resp, err
	}
}
