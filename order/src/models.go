package src

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// OrderItem struct.
// This struct is used to represent an order item.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
//   - Quantity (int64): The quantity.
type OrderItem struct {
	ProductID string `bson:"product_id" json:"product_id"`
	CompanyID  string `bson:"company_id" json:"company_id"`
	Quantity  int64  `bson:"quantity" json:"quantity"`
}

// Order struct.
// This struct is used to represent an order.
//
// Attributes:
//   - OrderID (primitive.ObjectID): The order ID.
//   - OrderItems ([]OrderItem): The order items.
//   - CustomerID (string): The customer ID.
//   - Status (string): The status.
//   - TotalPrice (float32): The total price.
//   - CreatedAt (time.Time): The created at timestamp.
//   - UpdatedAt (time.Time): The updated at timestamp.
type Order struct {
	OrderID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OrderItems    []OrderItem        `bson:"order_items" json:"order_items"`
	CustomerID    string             `bson:"customer_id" json:"customer_id"`
	Status        string             `bson:"status" json:"status"`
	TotalPrice    float32            `bson:"total_price" json:"total_price"`
	PaymentLinkID string             `bson:"payment_link_id" json:"payment_link_id"`
	PaymentID     string             `bson:"payment_id" json:"payment_id"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateOrderPayload struct.
// This struct is used to represent a request payload for creating an order.
//
// Attributes:
//   - OrderItems ([]OrderItem): The order items.
type CreateOrderPayload struct {
	OrderItems []OrderItem `json:"order_items"`
}

// CancelOrderPayload struct.
// This struct is used to represent a request payload for cancelling an order.
//
// Attributes:
//   - OrderID (string): The order ID.
//   - CustomerID (string): The customer ID.
type CancelOrderPayload struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
}

// SetOrderPaidPayload struct.
// This struct is used to represent a request payload for setting an order as paid.
//
// Attributes:
//   - OrderID (string): The order ID.
//   - CustomerID (string): The customer ID.
//   - PaymentID (string): The payment ID.
type SetOrderPaidPayload struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	PaymentID  string `json:"payment_id"`
}

// OrderCreatedEventPayloadOrderItem struct.
// This struct is used to represent an order item in the order created event payload.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
//   - PriceID (string): The product ID.
//   - Quantity (int64): The quantity.
type OrderCreatedEventPayloadOrderItem struct {
	ProductID string `json:"product_id"`
	CompanyID  string `json:"company_id"`
	PriceID   string `json:"price_id"`
	Quantity  int64  `json:"quantity"`
}

// OrderCancelledEventPayload struct.
// This struct is used to represent a payload for the order cancelled event.
//
// Attributes:
//   - OrderID (string): The order ID.
//   - CustomerID (string): The customer ID.
type OrderCancelledEventPayload struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
}

// OrderCreatedEventPayload struct.
// This struct is used to represent a payload for the order created event.
//
// Attributes:
type OrderCreatedEventPayload struct {
	OrderID    string                              `json:"order_id"`
	OrderItems []OrderCreatedEventPayloadOrderItem `json:"order_items"`
	CustomerID string                              `json:"customer_id"`
}

// OrderPaidSuccessEventPayload struct.
// This struct is used to represent a payload for the order paid success event.
//
// Attributes:
//   - OrderID (string): The order ID.
//   - CustomerID (string): The customer ID.
type OrderPaidSuccessEventPayload struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
}

type OrderValidator struct{
	PhonePattern string `json:"phone_pattern"`
}

// order id
// name 
// ph number
// house/flat name
// house/flat number
// post office
// district
// pincode
// payment type
// delay order
// status