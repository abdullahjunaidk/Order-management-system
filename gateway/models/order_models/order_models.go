package order_models

import "time"

// OrderItem struct.
// This struct is used to represent an item in an order.
type OrderItem struct {
	ProductID string `json:"product_id" validate:"required" example:"67adca0b4e864cfffb002299"`
	CompanyID  string `json:"company_id" validate:"required" example:"67adca0b4e864cfffb002299"`
	Quantity  int64  `json:"quantity" validate:"required" example:"1"`
}

type OrderItems struct{
	ProductID string
	CompanyID string
	Quantity int32
	Weight float32
}

// OrderCreatePayload struct.
// This struct is used to represent the payload for creating an order.
type OrderCreatePayload struct {
	Items []OrderItem `json:"items" validate:"required,dive,required"`
}

// Order struct.
// This struct is used to represent an order.
type Order struct {
	ID            string      `json:"id" example:"67adca0b4e864cfffb002299"`
	Items         []OrderItem `json:"items"`
	CustomerID    string      `json:"customer_id" example:"67adca0b4e864cfffb002299"`
	Status        string      `json:"status" example:"pending"`
	TotalPrice    float32     `json:"total_price" example:"100.00"`
	PaymentLinkID string      `json:"payment_link_id" example:"67adca0b4e864cfffb002299"`
	CreatedAt     time.Time   `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt     time.Time   `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// OrderCreateSuccessResponse struct.
// This struct is used to represent the success response for creating an order.
type OrderCreateSuccessResponse struct {
	Message string `json:"message" example:"Order Created Successfully!"`
	Order   Order  `json:"order"`
}

// OrderCreateErrorResponse struct.
// This struct is used to represent the error response for creating an order.
type OrderCreateErrorResponse struct {
	Message string `json:"message" example:"Failed to Create Order!"`
	Error   string `json:"error" example:"<error_message>"`
}

// GetOrderByIDSuccessResponse struct.
// This struct is used to represent the success response for getting an order by ID.
type GetOrderByIDSuccessResponse struct {
	Message string `json:"message" example:"Order Retrieved Successfully!"`
	Order   Order  `json:"order"`
}

// GetOrderByIDErrorResponse struct.
// This struct is used to represent the error response for getting an order by ID.
type GetOrderByIDErrorResponse struct {
	Message string `json:"message" example:"Failed to Retrieve Order!"`
	Error   string `json:"error" example:"<error_message>"`
}

// CancelOrderSuccessResponse struct.
// This struct is used to represent the success response for cancelling an order.
type CancelOrderSuccessResponse struct {
	Message string `json:"message" example:"Order Cancelled Successfully!"`
}

// CancelOrderErrorResponse struct.
// This struct is used to represent the error response for cancelling an order.
type CancelOrderErrorResponse struct {
	Message string `json:"message" example:"Failed to Cancel Order!"`
	Error   string `json:"error" example:"<error_message>"`
}

type OrderValidator struct{
	PhonePattern string `json:"phone_pattern"`
}