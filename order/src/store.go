package src

import (
	"context"
	"errors"

	mongoDatabase "common/database/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// MongoDB collection name.
	COLLECTION_ORDERS = "orders"
)

// OrderStore interface.
// This interface is used to define the order store methods.
//
// Methods:
//   - CreateOrder(ctx context.Context, order *Order) (string, error): Create a new order.
//   - GetOrderByIDAndCustomerID(ctx context.Context, orderID string, customerID string) (*Order, error): Get an order by ID and customer ID.
//   - SetOrderPaymentLinkID(ctx context.Context, orderID string, customerID string, paymentLinkID string) error: Set the payment link ID of an order.
//   - CancelOrder(ctx context.Context, orderID string, customerID string) error: Cancel an order.
//   - SetOrderPaid(ctx context.Context, orderID string, customerID string, paymentID string) error: Set an order to paid status.
type OrderStore interface {
	CreateOrder(ctx context.Context, order *Order) (string, error)
	GetOrderByIDAndCustomerID(ctx context.Context, orderID string, customerID string) (*Order, error)
	SetOrderPaymentLinkID(ctx context.Context, orderID string, customerID string, paymentLinkID string) error
	CancelOrder(ctx context.Context, orderID string, customerID string) error
	SetOrderPaid(ctx context.Context, orderID string, customerID string, paymentID string) error
}

// orderStore struct.
// This struct is used to implement the OrderStore interface.
//
// Attributes:
//   - ordersCollection (*mongo.Collection): The orders collection.
type orderStore struct {
	ordersCollection *mongo.Collection
}

// NewOrderStore function.
// This function is used to create a new order store.
//
// Parameters:
//   - adapter (mongoDatabase.MongoDBAdapter): The MongoDB adapter.
//
// Returns:
//   - OrderStore: The order store.
func NewOrderStore(adapter mongoDatabase.MongoDBAdapter) OrderStore {
	return &orderStore{
		ordersCollection: adapter.Collection(COLLECTION_ORDERS),
	}
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
func (s *orderStore) SetOrderPaymentLinkID(ctx context.Context, orderID string, customerID string, paymentLinkID string) error {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderStore.SetOrderPaymentLinkID")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	filter := map[string]interface{}{
		"_id":         objectID,
		"customer_id": customerID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"payment_link_id": paymentLinkID,
		},
	}

	_, err = s.ordersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	return nil
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
func (s *orderStore) CancelOrder(ctx context.Context, orderID string, customerID string) error {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderStore.CancelOrder")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	filter := map[string]interface{}{
		"_id":         objectID,
		"customer_id": customerID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status": "cancelled",
		},
	}

	result, err := s.ordersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	if result.ModifiedCount == 0 {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, errors.New("order not found or cannot be cancelled").Error())
		return errors.New("order not found or cannot be cancelled")
	}

	return nil
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
func (s *orderStore) SetOrderPaid(ctx context.Context, orderID string, customerID string, paymentID string) error {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderStore.SetOrderPaid")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	filter := map[string]interface{}{
		"_id":         objectID,
		"customer_id": customerID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"status":     "paid",
			"payment_id": paymentID,
		},
	}

	result, err := s.ordersCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	if result.ModifiedCount == 0 {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, errors.New("order not found or cannot be set to paid").Error())
		return errors.New("order not found or cannot be set to paid")
	}

	return nil
}

// CreateOrder method.
// This method is used to create a new order.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - order (*Order): The order.
//
// Returns:
//   - string: The order ID.
//   - error: An error if occurred.
func (s *orderStore) CreateOrder(ctx context.Context, order *Order) (string, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderStore.CreateOrder")
	defer span.End()

	res, err := s.ordersCollection.InsertOne(ctx, order)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return "", err
	}

	insertedID := res.InsertedID.(primitive.ObjectID).Hex()
	return insertedID, nil
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
func (s *orderStore) GetOrderByIDAndCustomerID(ctx context.Context, orderID string, customerID string) (*Order, error) {
	tracer := otel.Tracer("order-service")
	ctx, span := tracer.Start(ctx, "orderStore.GetOrderByIDAndCustomerID")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	filter := map[string]interface{}{
		"_id":         objectID,
		"customer_id": customerID,
	}

	var order Order
	err = s.ordersCollection.FindOne(ctx, filter).Decode(&order)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	return &order, nil
}
