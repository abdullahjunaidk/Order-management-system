package src

import (
	rabbitmqBroker "common/broker/rabbitmq"
	"common/helpers/env"
	"context"
	"errors"
	"time"

	inventorySrc "inventory/src"
	productSrc "product/src"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// RabbitMQ Queues.
	ORDER_CREATED_QUEUE           = env.GetEnv("ORDER_CREATED_QUEUE", "order.order.created")
	ORDER_PENDING_CANCELLED_QUEUE = env.GetEnv("ORDER_PENDING_CANCELLED_QUEUE", "order.pending.cancelled")
	ORDER_PAID_SUCCESS_QUEUE      = env.GetEnv("ORDER_PAID_SUCCESS_QUEUE", "order.paid.success")
	ORDER_PAID_CANCELLED_QUEUE    = env.GetEnv("ORDER_PAID_CANCELLED_QUEUE", "order.paid.cancelled")
)

// OrderService interface.
// This interface is used to define the order service methods.
//
// Methods:
//   - CreateOrder(ctx context.Context, payload CreateOrderPayload, customerID string) (*Order, error): Create a new order.
//   - GetOrderByIDAndCustomerID(ctx context.Context, orderID string, customerID string) (*Order, error): Get an order by ID and customer ID.
//   - SetOrderPaymentLinkID(ctx context.Context, orderID string, customerID string, paymentLinkID string) error: Set the payment link ID of an order.
//   - CancelOrder(ctx context.Context, orderID string, customerID string) error: Cancel an order.
//   - SetOrderPaid(ctx context.Context, orderID string, customerID string, paymentID string) error: Set an order to paid status.
type OrderService interface {
	CreateOrder(ctx context.Context, payload CreateOrderPayload, customerID string) (*Order, error)
	GetOrderByIDAndCustomerID(ctx context.Context, orderID string, customerID string) (*Order, error)
	SetOrderPaymentLinkID(ctx context.Context, orderID string, customerID string, paymentLinkID string) error
	CancelOrder(ctx context.Context, orderID string, customerID string) error
	SetOrderPaid(ctx context.Context, orderID string, customerID string, paymentID string) error
}

// orderService struct.
// This struct is used to implement the OrderService interface.
//
// Attributes:
//   - store (OrderStore): The order store.
//   - productClient (*productSrc.ProductClient): The product client.
//   - inventoryClient (*inventorySrc.InventoryClient): The inventory client.
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
type orderService struct {
	store           OrderStore
	productClient   *productSrc.ProductClient
	inventoryClient *inventorySrc.InventoryClient
	rabbitMQ        *rabbitmqBroker.RabbitMQAdapter
}

// NewOrderService function.
// This function is used to create a new order service.
//
// Parameters:
//   - store (OrderStore): The order store.
//   - productClient (*productSrc.ProductClient): The product client.
//   - inventoryClient (*inventorySrc.InventoryClient): The inventory client.
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
//
// Returns:
//   - OrderService: The order service.
func NewOrderService(store OrderStore, productClient *productSrc.ProductClient, inventoryClient *inventorySrc.InventoryClient, rabbitMQ *rabbitmqBroker.RabbitMQAdapter) OrderService {
	return &orderService{
		store:           store,
		productClient:   productClient,
		inventoryClient: inventoryClient,
		rabbitMQ:        rabbitMQ,
	}
}

// CreateOrder method.
// This method is used to create a new order.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - payload (CreateOrderPayload): The order payload.
//   - customerID (string): The customer ID.
//
// Returns:
//   - *Order: The created order.
//   - error: An error if occurred.
func (s *orderService) CreateOrder(ctx context.Context, payload CreateOrderPayload, customerID string) (*Order, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderService.CreateOrder")
	defer span.End()

	var total float32 = 0

	orderItems := make([]OrderItem, len(payload.OrderItems))
	eventPayloadOrderItems := make([]OrderCreatedEventPayloadOrderItem, len(payload.OrderItems))
	for i, item := range payload.OrderItems {
		product, err := s.productClient.GetProductByIDAndCompanyID(ctx, item.ProductID, item.CompanyID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return nil, err
		}

		inventory, err := s.inventoryClient.GetInventoryByProductIDAndCompanyID(ctx, item.ProductID, item.CompanyID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return nil, err
		}

		if inventory.AvailableQuantity < item.Quantity {
			err := errors.New("insufficient quantity")
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return nil, err
		}

		orderItems[i] = OrderItem{
			ProductID: product.Id,
			CompanyID: product.CompanyId,
			Quantity:  item.Quantity,
		}

		eventPayloadOrderItems[i] = OrderCreatedEventPayloadOrderItem{
			ProductID: product.Id,
			CompanyID: product.CompanyId,
			PriceID:   product.PriceId,
			Quantity:  item.Quantity,
		}

		total += product.Price * float32(item.Quantity)
	}

	order := &Order{
		OrderID:       primitive.NewObjectID(),
		OrderItems:    orderItems,
		CustomerID:    customerID,
		Status:        "pending",
		TotalPrice:    total,
		PaymentLinkID: "",
		PaymentID:     "",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	orderID, err := s.store.CreateOrder(ctx, order)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	createdOrder, err := s.store.GetOrderByIDAndCustomerID(ctx, orderID, customerID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	eventPayload := OrderCreatedEventPayload{
		OrderID:    createdOrder.OrderID.Hex(),
		OrderItems: eventPayloadOrderItems,
		CustomerID: createdOrder.CustomerID,
	}

	if err := s.rabbitMQ.PublishMessage(ctx, ORDER_CREATED_QUEUE, eventPayload); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	for _, item := range payload.OrderItems {
		inventory, err := s.inventoryClient.GetInventoryByProductIDAndCompanyID(ctx, item.ProductID, item.CompanyID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return nil, err
		}

		inventory.AvailableQuantity -= item.Quantity

		_, err = s.inventoryClient.UpdateInventory(ctx, inventory.ProductId, inventory.CompanyId, inventory.AvailableQuantity, inventory.ThresholdQuantity)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return nil, err
		}
	}

	span.SetStatus(otlpcodes.Ok, "Order Created Successfully!")
	return createdOrder, nil
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
//   - error: An error if occurred.
func (s *orderService) CancelOrder(ctx context.Context, orderID string, customerID string) error {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderService.CancelOrder")
	defer span.End()

	order, err := s.store.GetOrderByIDAndCustomerID(ctx, orderID, customerID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	err = s.store.CancelOrder(ctx, orderID, customerID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	for i, item := range order.OrderItems {
		inventory, err := s.inventoryClient.GetInventoryByProductIDAndCompanyID(ctx, item.ProductID, item.CompanyID)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return err
		}

		inventory.AvailableQuantity += order.OrderItems[i].Quantity

		_, err = s.inventoryClient.UpdateInventory(ctx, inventory.ProductId, inventory.CompanyId, inventory.AvailableQuantity, inventory.ThresholdQuantity)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return err
		}
	}

	eventPayload := OrderCancelledEventPayload{
		OrderID:    orderID,
		CustomerID: customerID,
	}

	var queueName string
	switch order.Status {
	case "pending":
		queueName = ORDER_PENDING_CANCELLED_QUEUE
	case "paid":
		queueName = ORDER_PAID_CANCELLED_QUEUE
	default:
		err := errors.New("invalid order status")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	if err := s.rabbitMQ.PublishMessage(ctx, queueName, eventPayload); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Order Cancelled Successfully and event published")
	return nil
}

// SetOrderPaymentLinkID method.
// This method is used to set the payment link ID of an order.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - orderID (string): The order ID.
//   - customerID (string): The customer ID.
//   - paymentLinkID (string): The payment link ID.
//
// Returns:
//   - error: An error if occurred.
func (s *orderService) SetOrderPaymentLinkID(ctx context.Context, orderID string, customerID string, paymentLinkID string) error {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderService.SetOrderPaymentLinkID")
	defer span.End()

	err := s.store.SetOrderPaymentLinkID(ctx, orderID, customerID, paymentLinkID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Order Payment Link ID Set Successfully!")
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
//   - *Order: The order.
//   - error: An error if occurred.
func (s *orderService) GetOrderByIDAndCustomerID(ctx context.Context, orderID string, customerID string) (*Order, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderService.GetOrderByIDAndCustomerID")
	defer span.End()

	order, err := s.store.GetOrderByIDAndCustomerID(ctx, orderID, customerID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Order Fetched Successfully!")
	return order, nil
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
//   - error: An error if occurred.
func (s *orderService) SetOrderPaid(ctx context.Context, orderID string, customerID string, paymentID string) error {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderService.SetOrderPaid")
	defer span.End()

	err := s.store.SetOrderPaid(ctx, orderID, customerID, paymentID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	eventPayload := OrderPaidSuccessEventPayload{
		OrderID:    orderID,
		CustomerID: customerID,
	}

	if err := s.rabbitMQ.PublishMessage(ctx, ORDER_PAID_SUCCESS_QUEUE, eventPayload); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Order Paid Successfully!")
	return nil
}
