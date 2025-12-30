package src

// ProductCreatedEventPayload struct.
// This struct is used to represent a product created event payload.
//
// Attributes:
//   - ID (string): The product ID.
//   - VendorID (string): The vendor ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
type ProductCreatedEventPayload struct {
	ID          string  `json:"id"`
	VendorID    string  `json:"vendor_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float32 `json:"price"`
}

// ProductUpdatedEventPayload struct.
// This struct is used to represent a product updated event payload.
//
// Attributes:
//   - ID (string): The product ID.
//   - VendorID (string): The vendor ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
//   - PriceID (string): The product price ID.
type ProductUpdatedEventPayload struct {
	ID          string  `json:"id"`
	VendorID    string  `json:"vendor_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float32 `json:"price"`
	PriceID     string  `json:"price_id"`
}

// ProductDeletedEventPayload struct.
// This struct is used to represent a product delete payload.
//
// Attributes:
//   - ID (string): The product ID.
type ProductDeletedEventPayload struct {
	ID string `json:"id"`
}

// OrderCreatedEventPayloadOrderItem struct.
// This struct is used to represent an order item in the order created event payload.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - VendorID (string): The vendor ID.
//   - PriceID (string): The product ID.
//   - Quantity (int64): The quantity.
type OrderCreatedEventPayloadOrderItem struct {
	ProductID string `json:"product_id"`
	VendorID  string `json:"vendor_id"`
	PriceID   string `json:"price_id"`
	Quantity  int64  `json:"quantity"`
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
