package src

import (
	"common/helpers/logger"
	orderProto "common/proto/order"
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrderClient struct.
// This struct is used to create a new client connection to the order service.
//
// Attributes:
//   - conn (*grpc.ClientConn): The client connection to the order service.
//   - service (orderProto.OrderServiceClient): The order service client.
//   - tracerName (string): The tracer name.
//   - log (*logrus.Logger): The logger.
type OrderClient struct {
	conn       *grpc.ClientConn
	service    orderProto.OrderServiceClient
	tracerName string
	log        *logrus.Logger
}

// NewOrderClient function.
// This function is used to create a new client connection to the order service.
//
// Parameters:
//   - address (string): The address of the order service.
//
// Returns:
//   - *OrderClient: The order client.
//   - error: The error.
func NewOrderClient(address string, tracerName string) (*OrderClient, error) {
	log := logger.NewLogger()

	clientOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, clientOpts...)
	if err != nil {
		log.WithFields(logrus.Fields{"address": address, "error": err}).Error("Failed to Create Client Connection!")
		return nil, fmt.Errorf("failed to create client connection: %w", err)
	}

	service := orderProto.NewOrderServiceClient(conn)

	return &OrderClient{
		conn:       conn,
		service:    service,
		tracerName: tracerName,
		log:        log,
	}, nil
}

// Close function.
// This function is used to close the client connection to the order service.
func (c *OrderClient) Close() {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			c.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close Connection!")
		} else {
			c.log.Info("Connection Closed!")
		}
	}
}

// CreateOrder method.
// This method is used to create a new order.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - items ([]*orderProto.OrderItem): The order items.
//   - customerID (string): The customer ID.
//
// Returns:
//   - *orderProto.Order: The created order.
//   - error: The error.
func (c *OrderClient) CreateOrder(ctx context.Context, items []*orderProto.OrderItem, customerID string) (*orderProto.Order, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "OrderClient.CreateOrder")
	defer span.End()

	req := &orderProto.OrderCreatePayload{
		Items:      items,
		CustomerId: customerID,
	}

	res, err := c.service.CreateOrder(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"customer_id": customerID, "error": err}).Error("Failed to Create Order!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"customer_id": customerID}).Info("Order Created Successfully!")
	return res, nil
}

// CancelOrder method.
// This method is used to cancel an order.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - orderID (string): The order ID.
//   - customerID (string): The customer ID.
//
// Returns:
//   - error: The error.
func (c *OrderClient) CancelOrder(ctx context.Context, orderID string, customerID string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "OrderClient.CancelOrder")
	defer span.End()

	req := &orderProto.CancelOrderPayload{
		OrderId:    orderID,
		CustomerId: customerID,
	}

	_, err := c.service.CancelOrder(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"order_id": orderID, "customer_id": customerID, "error": err}).Error("Failed to Cancel Order!")
		return err
	}

	c.log.WithFields(logrus.Fields{"order_id": orderID, "customer_id": customerID}).Info("Order Cancelled Successfully!")
	return nil
}

// SetOrderPaymentLinkID method.
// This method is used to get an order by ID and customer ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - orderID (string): The order ID.
//   - customerID (string): The customer ID.
//
// Returns:
//   - *orderProto.Order: The order.
//   - error: The error.
func (c *OrderClient) SetOrderPaymentLinkID(ctx context.Context, orderID string, customerID string, paymentLinkID string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "OrderClient.SetOrderPaymentLinkID")
	defer span.End()

	req := &orderProto.SetOrderPaymentLinkIDPayload{
		OrderId:       orderID,
		CustomerId:    customerID,
		PaymentLinkId: paymentLinkID,
	}

	_, err := c.service.SetOrderPaymentLinkID(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"order_id": orderID, "customer_id": customerID, "payment_link_id": paymentLinkID, "error": err}).Error("Failed to Set Order Payment Link ID!")
		return err
	}

	c.log.WithFields(logrus.Fields{"order_id": orderID, "customer_id": customerID, "payment_link_id": paymentLinkID}).Info("Order Payment Link ID Set Successfully!")
	return nil
}

// GetOrderByIDAndCustomerID method.
// This method is used to get an order by ID and customer ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - orderID (string): The order ID.
//   - customerID (string): The customer ID.
//
// Returns:
//   - *orderProto.Order: The order.
//   - error: The error.
func (c *OrderClient) GetOrderByIDAndCustomerID(ctx context.Context, orderID string, customerID string) (*orderProto.Order, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "OrderClient.GetOrderByIDAndCustomerID")
	defer span.End()

	req := &orderProto.GetOrderByIDAndCustomerIDPayload{
		OrderId:    orderID,
		CustomerId: customerID,
	}

	res, err := c.service.GetOrderByIDAndCustomerID(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"order_id": orderID, "customer_id": customerID, "error": err}).Error("Failed to Get Order!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"order_id": orderID, "customer_id": customerID}).Info("Order Fetched Successfully!")
	return res, nil
}

// SetOrderPaid method.
// This method is used to set an order to paid status.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - orderID (string): The order ID.
//   - customerID (string): The customer ID.
//   - paymentID (string): The payment ID.
//
// Returns:
//   - error: The error.
func (c *OrderClient) SetOrderPaid(ctx context.Context, orderID string, customerID string, paymentID string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "OrderClient.SetOrderPaid")
	defer span.End()

	req := &orderProto.SetOrderPaidPayload{
		OrderId:    orderID,
		CustomerId: customerID,
		PaymentId:  paymentID,
	}

	_, err := c.service.SetOrderPaid(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"order_id": orderID, "customer_id": customerID, "error": err}).Error("Failed to Set Order Paid!")
		return err
	}

	c.log.WithFields(logrus.Fields{"order_id": orderID, "customer_id": customerID}).Info("Order Paid Successfully!")
	return nil
}
