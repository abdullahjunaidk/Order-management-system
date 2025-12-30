package inventory_models

import "time"

// Inventory struct.
// This struct is used to represent an Inventory.
//
// Attributes:
//   - ID (string): The inventory ID.
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
//   - AvailableQuantity (int64): The available quantity.
//   - ThresholdQuantity (int64): The threshold quantity.
//   - CreatedAt (time.Time): The creation time.
//   - UpdatedAt (time.Time): The update time.
type Inventory struct {
	ID                string    `json:"id,omitempty" example:"67adca0b4e864cfffb002299"`
	ProductID         string    `json:"product_id,omitempty" example:"67adca0b4e864cfffb002299"`
	CompanyID         string    `json:"company_id,omitempty" example:"67adca0b4e864cfffb002299"`
	AvailableQuantity int64     `json:"available_quantity,omitempty" example:"100"`
	ThresholdQuantity int64     `json:"threshold_quantity,omitempty" example:"10"`
	CreatedAt         time.Time `json:"created_at,omitempty" example:"2021-01-01T00:00:00Z"`
	UpdatedAt         time.Time `json:"updated_at,omitempty" example:"2021-01-01T00:00:00Z"`
}

// CreateInventoryPayload struct.
// This struct is used to represent a request payload for creating inventory.
//
// Attributes:
//   - AvailableQuantity (int64): The available quantity.
//   - ThresholdQuantity (int64): The threshold quantity.
type CreateInventoryPayload struct {
	AvailableQuantity int64 `json:"available_quantity" validate:"required" example:"100"`
	ThresholdQuantity int64 `json:"threshold_quantity" validate:"required" example:"10"`
}

// CreateInventorySuccessResponse struct.
// This struct is used to represent a success response for creating inventory.
//
// Attributes:
//   - Message (string): Success message.
//   - Inventory (Inventory): Created inventory.
type CreateInventorySuccessResponse struct {
	Message   string    `json:"message" example:"Inventory Created Successfully!"`
	Inventory Inventory `json:"inventory"`
}

// CreateInventoryErrorResponse struct.
// This struct is used to represent an error response for creating inventory.
//
// Attributes:
//   - Message (string): Error message.
//   - Error (string): Detailed error string.
type CreateInventoryErrorResponse struct {
	Message string `json:"message" example:"Failed to Create Inventory!"`
	Error   string `json:"error" example:"<error_message>"`
}

// GetInventoryByProductIDAndVendorIDPayload struct.
// This struct is used to represent a request payload for getting inventory by product and company IDs.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
type GetInventoryByProductIDAndCompanyIDPayload struct {
	ProductID string `json:"product_id" validate:"required" example:"67adca0b4e864cfffb002299"`
}

// GetInventoryByProductIDAndCompanyIDSuccessResponse struct.
// This struct is used to represent a success response for getting inventory by product and company IDs.
//
// Attributes:
//   - Message (string): Success message.
//   - Inventory (Inventory): Retrieved inventory.
type GetInventoryByProductIDAndCompanyIDSuccessResponse struct {
	Message   string    `json:"message" example:"Inventory Fetched Successfully!"`
	Inventory Inventory `json:"inventory"`
}

// GetInventoryByProductIDAndCompanyIDErrorResponse struct.
// This struct is used to represent an error response for getting inventory by product and company IDs.
//
// Attributes:
//   - Message (string): Error message.
//   - Error (string): Detailed error string.
type GetInventoryByProductIDAndCompanyIDErrorResponse struct {
	Message string `json:"message" example:"Failed to Get Inventory!"`
	Error   string `json:"error" example:"<error_message>"`
}

// DeleteInventoryByProductIDAndCompanyIDPayload struct.
// This struct is used to represent a request payload for deleting inventory by product and company IDs.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
type DeleteInventoryByProductIDAndCompanyIDPayload struct {
	ProductID string `json:"product_id" validate:"required" example:"67adca0b4e864cfffb002299"`
}

// DeleteInventoryByProductIDAndCompanyIDSuccessResponse struct.
// This struct is used to represent a success response for deleting inventory by product and company IDs.
//
// Attributes:
//   - Message (string): Success message.
type DeleteInventoryByProductIDAndCompanyIDSuccessResponse struct {
	Message string `json:"message" example:"Inventory Deleted Successfully!"`
}

// DeleteInventoryByProductIDAndCompanyIDErrorResponse struct.
// This struct is used to represent an error response for deleting inventory by product and company IDs.
//
// Attributes:
//   - Message (string): Error message.
//   - Error (string): Detailed error string.
type DeleteInventoryByProductIDAndCompanyIDErrorResponse struct {
	Message string `json:"message" example:"Failed to Delete Inventory!"`
	Error   string `json:"error" example:"<error_message>"`
}

// UpdateInventoryPayload struct.
// This struct is used to represent a request payload for updating inventory.
//
// Attributes:
//   - AvailableQuantity (int64): The available quantity.
//   - ThresholdQuantity (int64): The threshold quantity.
type UpdateInventoryPayload struct {
	AvailableQuantity int64 `json:"available_quantity" validate:"required" example:"100"`
	ThresholdQuantity int64 `json:"threshold_quantity" validate:"required" example:"10"`
}

// UpdateInventorySuccessResponse struct.
// This struct is used to represent a success response for updating inventory.
//
// Attributes:
//   - Message (string): Success message.
//   - Inventory (Inventory): Updated inventory.
type UpdateInventorySuccessResponse struct {
	Message   string    `json:"message" example:"Inventory Updated Successfully!"`
	Inventory Inventory `json:"inventory"`
}

// UpdateInventoryErrorResponse struct.
// This struct is used to represent an error response for updating inventory.
//
// Attributes:
//   - Message (string): Error message.
//   - Error (string): Detailed error string.
type UpdateInventoryErrorResponse struct {
	Message string `json:"message" example:"Failed to Update Inventory!"`
	Error   string `json:"error" example:"<error_message>"`
}
