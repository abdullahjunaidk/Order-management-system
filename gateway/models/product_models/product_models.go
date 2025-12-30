package product_models

import "time"

// Product struct.
// This struct is used to represent a product.
//
// Attributes:
//   - ID (string): The product ID.
//   - CompanyID (string): The company ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
//   - CreatedAt (time.Time): The product creation date.
//   - UpdatedAt (time.Time): The product update date.
type Product struct {
	ID          string            `json:"id,omitempty" example:"67adca0b4e864cfffb002299"`
	CompanyID   string            `json:"company_id,omitempty" example:"67adca0b4e864cfffb002299"`
	Name        string            `json:"name,omitempty" example:"Product Name"`
	Description string            `json:"description,omitempty" example:"Product Description"`
	Category    string            `json:"category,omitempty" example:"Product Category"`
	Price       float32           `json:"price,omitempty" example:"100.00"`
	NetWeight   float32           `json:"net_weight,omitempty" example:"100.00"`
	CreatedAt   time.Time         `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt   time.Time         `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	Inventory   *ProductInventory `json:"inventory,omitempty"`
}

// ProductInventory struct.
// This struct is used to represent a product inventory.
//
// Attributes:
//   - AvailableQuantity (int64): The available quantity.
//   - ThresholdQuantity (int64): The threshold quantity.
//   - CreatedAt (time.Time): The creation time.
//   - UpdatedAt (time.Time): The update time.
type ProductInventory struct {
	AvailableQuantity int64     `json:"available_quantity,omitempty" example:"100"`
	ThresholdQuantity int64     `json:"threshold_quantity,omitempty" example:"10"`
	CreatedAt         time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt         time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

// ProductCreatePayload struct.
// This struct is used to represent a product create payload.
//
// Attributes:
//   - CompanyID (string): The company ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
//   - NetWeight (float32): The product net weight.
type ProductCreatePayload struct {
	CompanyID   string  `json:"company_id,omitempty" example:"67adca0b4e864cfffb002299" validate:"required"`
	Name        string  `json:"name,omitempty" example:"Product Name" validate:"required"`
	Description string  `json:"description,omitempty" example:"Product Description" validate:"required"`
	Category    string  `json:"category,omitempty" example:"Product Category"`
	Price       float32 `json:"price,omitempty" example:"100.00" validate:"required"`
	NetWeight   float32 `json:"net_weight,omitempty" example:"100.00" validate:"required"`
}

// ProductCreateSuccessResponse struct.
// This struct is used to represent a product create response.
//
// Attributes:
//   - Message (string): The response message.
//   - Product (Product): The created product.
type ProductCreateSuccessResponse struct {
	Message string  `json:"message" example:"Product Created Successfully!"`
	Product Product `json:"product"`
}

// ProductCreateErrorResponse struct.
// This struct is used to represent a product create error response.
//
// Attributes:
//   - Message (string): The response message.
//   - Error (string): The error message.
type ProductCreateErrorResponse struct {
	Message string `json:"message" example:"Failed to Create Product!"`
	Error   string `json:"error" example:"<error_message>"`
}

// ProductUpdatePayload struct.
// This struct is used to represent a product update payload.
//
// Attributes:
//   - CompanyID (string): The company ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
type ProductUpdatePayload struct {
	CompanyID   string  `json:"company_id,omitempty" example:"67adca0b4e864cfffb002299" validate:"required"`
	Name        string  `json:"name,omitempty" example:"Product Name" validate:"required"`
	Description string  `json:"description,omitempty" example:"Product Description" validate:"required"`
	Category    string  `json:"category,omitempty" example:"Product Category"`
	Price       float32 `json:"price,omitempty" example:"100.00" validate:"required"`
	NetWeight   float32 `json:"net_weight,omitempty" example:"100.00" validate:"required"`
}

// ProductUpdateSuccessResponse struct.
// This struct is used to represent a product update success response.
//
// Attributes:
//   - Message (string): The response message.
//   - Product (Product): The updated product.
type ProductUpdateSuccessResponse struct {
	Message string  `json:"message" example:"Product Updated Successfully!"`
	Product Product `json:"product"`
}

// ProductUpdateErrorResponse struct.
// This struct is used to represent a product update error response.
//
// Attributes:
//   - Message (string): The response message.
//   - Error (string): The error message.
type ProductUpdateErrorResponse struct {
	Message string `json:"message" example:"Failed to Update Product!"`
	Error   string `json:"error" example:"<error_message>"`
}

// ListProductsByCompanyIDPayload struct.
// This struct is used to represent list products by company id payload.
//
// Attributes:
//   - Limit (int64): The limit of products to return.
//   - Offset (int64): The offset of products to return.
type ListProductsByCompanyIDPayload struct {
	CompanyID string `form:"company_id" json:"company_id" validate:"required" example:"67adca0b4e864cfffb002299"`
	Limit     int64  `form:"limit,default=10" json:"limit,omitempty" example:"10"`
	Offset    int64  `form:"offset,default=0" json:"offset,omitempty" example:"0"`
}

// ListProductsByCompanyIDSuccessResponse struct.
// This struct is used to represent list products by company id success response.
//
// Attributes:
//   - Message (string): The response message.
//   - Products ([]*Product): The list of products.
//   - TotalCount (int64): The total count of products.
type ListProductsByCompanyIDSuccessResponse struct {
	Message    string    `json:"message" example:"Products Listed Successfully!"`
	Products   []Product `json:"products"`
	TotalCount int64     `json:"total_count" example:"100"`
}

// ListProductsByCompanyIDErrorResponse struct.
// This struct is used to represent list products by company id error response.
//
// Attributes:
//   - Message (string): The response message.
//   - Error (string): The error message.
type ListProductsByCompanyIDErrorResponse struct {
	Message string `json:"message" example:"Failed to List Products!"`
	Error   string `json:"error" example:"<error_message>"`
}

// DeleteProductPayload struct.
// This struct is used to represent delete product payload.
//
// Attributes:
//   - ID (string): The product ID.
type DeleteProductPayload struct {
	ID string `json:"id" validate:"required" example:"67adca0b4e864cfffb002299"`
}

// DeleteProductSuccessResponse struct.
// This struct is used to represent delete product success response.
//
// Attributes:
//   - Message (string): The response message.
type DeleteProductSuccessResponse struct {
	Message string `json:"message" example:"Product Deleted Successfully!"`
}

// DeleteProductErrorResponse struct.
// This struct is used to represent delete product error response.
//
// Attributes:
//   - Message (string): The response message.
//   - Error (string): The error message.
type DeleteProductErrorResponse struct {
	Message string `json:"message" example:"Failed to Delete Product!"`
	Error   string `json:"error" example:"<error_message>"`
}
