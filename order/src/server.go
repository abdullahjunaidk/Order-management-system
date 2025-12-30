package src

import (
	"common/helpers/logger"
	orderProto "common/proto/order"
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// orderGRPCServer struct.
// This struct is used to implement the order service gRPC server.
//
// Attributes:
//   - orderService (OrderService): The order service.
//   - log (*logrus.Logger): The logger.
type orderGRPCServer struct {
	orderProto.UnimplementedOrderServiceServer
	orderService OrderService
	log          *logrus.Logger
}

// ListenAndServeGRPC function.
// This function is used to listen and serve the order service gRPC server.
//
// Parameters:
//   - service (OrderService): The order service.
//   - port (int): The port.
//
// Returns:
//   - error: The error.
func ListenAndServeGRPC(service OrderService, port int) error {
	log := logger.NewLogger()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithFields(logrus.Fields{"port": port, "error": err}).Error("Failed to Listen!")
		return fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()

	orderProto.RegisterOrderServiceServer(server, &orderGRPCServer{orderService: service, log: log})
	grpc_health_v1.RegisterHealthServer(server, &healthServer{})
	reflection.Register(server)

	log.WithFields(logrus.Fields{"port": port}).Info("Starting gRPC Server...")
	if err := server.Serve(lis); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Serve gRPC Server!")
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// CreateOrder method.
// This method is used to create a new order.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*orderProto.OrderCreatePayload): The create order request.
//
// Returns:
//   - *orderProto.Order: The created order.
//   - error: The error.
func (s *orderGRPCServer) CreateOrder(ctx context.Context, req *orderProto.OrderCreatePayload) (*orderProto.Order, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderGRPCServer.CreateOrder")
	defer span.End()

	s.log.WithFields(logrus.Fields{"customer_id": req.CustomerId}).Info("Creating New Order...")

	orderItems := make([]OrderItem, len(req.Items))
	for i, item := range req.Items {
		orderItems[i] = OrderItem{
			ProductID: item.ProductId,
			CompanyID: item.CompanyId,
			Quantity:  item.Quantity,
		}
	}

	payload := CreateOrderPayload{
		OrderItems: orderItems,
	}

	res, err := s.orderService.CreateOrder(ctx, payload, req.CustomerId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"customer_id": req.CustomerId, "error": err}).Error("Failed to Create Order!")
		return nil, err
	}

	items := make([]*orderProto.OrderItem, len(res.OrderItems))
	for i, item := range res.OrderItems {
		items[i] = &orderProto.OrderItem{
			ProductId: item.ProductID,
			CompanyId: item.CompanyID,
			Quantity:  item.Quantity,
		}
	}

	s.log.WithFields(logrus.Fields{"customer_id": req.CustomerId}).Info("Order Created Successfully!")
	return &orderProto.Order{
		OrderId:       res.OrderID.Hex(),
		Items:         items,
		CustomerId:    res.CustomerID,
		Status:        res.Status,
		TotalPrice:    res.TotalPrice,
		PaymentLinkId: res.PaymentLinkID,
		PaymentId:     res.PaymentID,
		CreatedAt:     timestamppb.New(res.CreatedAt),
		UpdatedAt:     timestamppb.New(res.UpdatedAt),
	}, nil
}

// CancelOrder method.
// This method is used to cancel an order.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*orderProto.CancelOrderPayload): The cancel order request.
//
// Returns:
//   - *emptypb.Empty: The empty response.
//   - error: The error.
func (s *orderGRPCServer) CancelOrder(ctx context.Context, req *orderProto.CancelOrderPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderGRPCServer.CancelOrder")
	defer span.End()

	s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId}).Info("Cancelling Order...")

	err := s.orderService.CancelOrder(ctx, req.OrderId, req.CustomerId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId, "error": err}).Error("Failed to Cancel Order!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId}).Info("Order Cancelled Successfully!")
	return &emptypb.Empty{}, nil
}

// SetOrderPaymentLinkID method.
// This method is used to set the payment link ID of an order.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*orderProto.SetOrderPaymentLinkIDPayload): The set order payment link ID request.
//
// Returns:
//   - *emptypb.Empty: The empty response.
//   - error: The error.
func (s *orderGRPCServer) SetOrderPaymentLinkID(ctx context.Context, req *orderProto.SetOrderPaymentLinkIDPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderGRPCServer.SetOrderPaymentLinkID")
	defer span.End()

	s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId, "payment_link_id": req.PaymentLinkId}).Info("Setting Order Payment Link ID...")

	err := s.orderService.SetOrderPaymentLinkID(ctx, req.OrderId, req.CustomerId, req.PaymentLinkId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId, "payment_link_id": req.PaymentLinkId, "error": err}).Error("Failed to Set Order Payment Link ID!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId, "payment_link_id": req.PaymentLinkId}).Info("Order Payment Link ID Set Successfully!")
	return &emptypb.Empty{}, nil
}

// GetOrderByIDAndCustomerID method.
// This method is used to get an order by ID and customer ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*orderProto.GetOrderByIDAndCustomerIDPayload): The get order by ID and customer ID request.
//
// Returns:
//   - *orderProto.Order: The order.
//   - error: The error.
func (s *orderGRPCServer) GetOrderByIDAndCustomerID(ctx context.Context, req *orderProto.GetOrderByIDAndCustomerIDPayload) (*orderProto.Order, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderGRPCServer.GetOrderByIDAndCustomerID")
	defer span.End()

	s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId}).Info("Getting Order...")

	res, err := s.orderService.GetOrderByIDAndCustomerID(ctx, req.OrderId, req.CustomerId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId, "error": err}).Error("Failed to Get Order!")
		return nil, err
	}

	items := make([]*orderProto.OrderItem, len(res.OrderItems))
	for i, item := range res.OrderItems {
		items[i] = &orderProto.OrderItem{
			ProductId: item.ProductID,
			CompanyId: item.CompanyID,
			Quantity:  item.Quantity,
		}
	}

	s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId}).Info("Order Fetched Successfully!")
	return &orderProto.Order{
		OrderId:       res.OrderID.Hex(),
		Items:         items,
		CustomerId:    res.CustomerID,
		Status:        res.Status,
		TotalPrice:    res.TotalPrice,
		PaymentLinkId: res.PaymentLinkID,
		PaymentId:     res.PaymentID,
		CreatedAt:     timestamppb.New(res.CreatedAt),
		UpdatedAt:     timestamppb.New(res.UpdatedAt),
	}, nil
}

// SetOrderPaid method.
// This method is used to set an order to paid status.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*orderProto.SetOrderPaidPayload): The set order paid request.
//
// Returns:
//   - *emptypb.Empty: The empty response.
//   - error: The error.
func (s *orderGRPCServer) SetOrderPaid(ctx context.Context, req *orderProto.SetOrderPaidPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderGRPCServer.SetOrderPaid")
	defer span.End()

	s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId, "payment_id": req.PaymentId}).Info("Setting Order to Paid...")

	err := s.orderService.SetOrderPaid(ctx, req.OrderId, req.CustomerId, req.PaymentId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId, "payment_id": req.PaymentId, "error": err}).Error("Failed to Set Order to Paid!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"order_id": req.OrderId, "customer_id": req.CustomerId, "payment_id": req.PaymentId}).Info("Order Set to Paid Successfully!")
	return &emptypb.Empty{}, nil
}
