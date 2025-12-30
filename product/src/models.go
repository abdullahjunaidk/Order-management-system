package src

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product struct.
// This struct is used to represent a product.
//
// Attributes:
//   - ID (primitive.ObjectID): The product ID.
//   - CompanyID (string): The company ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
//   - PriceID (string): The product price ID.
//   - CreatedAt (time.Time): The product creation date.
//   - UpdatedAt (time.Time): The product last update date.
type Product struct {
	BaseModel `bson:",inline" json:",inline"`
	CompanyID string             `bson:"company_id" json:"company_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Category    string             `bson:"category" json:"category"`
	Price       float32            `bson:"price" json:"price"`
	PriceID     string             `bson:"price_id" json:"price_id"`
	NetWeight   float32            `bson:"net_weight" json:"net_weight"`
}

// ProductCreatePayload struct.
// This struct is used to represent a product creation payload.
//
// Attributes:
//   - CompanyID (string): The company ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Price (float32): The product price.
type ProductCreatePayload struct {
	CompanyID   string  `json:"company_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float32 `json:"price"`
}

// ProductCreatedEventPayload struct.
// This struct is used to represent a product created event payload.
//
// Attributes:
//   - ID (string): The product ID.
//   - CompanyID (string): The company ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
type ProductCreatedEventPayload struct {
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float32 `json:"price"`
}

// ProductUpdatePayload struct.
// This struct is used to represent a product update payload.
//
// Attributes:
//   - ID (string): The product ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
type ProductUpdatePayload struct {
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
//   - CompanyID (string): The company ID.
//   - Name (string): The product name.
//   - Description (string): The product description.
//   - Category (string): The product category.
//   - Price (float32): The product price.
//   - PriceID (string): The product price ID.
type ProductUpdatedEventPayload struct {
	ID          string  `json:"id"`
	CompanyID   string  `json:"company_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float32 `json:"price"`
	PriceID     string  `json:"price_id"`
}

// ListProductsByCompanyIDPayload struct.
// This struct is used to represent a list products by company ID payload.
//
// Attributes:
//   - CompanyID (string): The company ID.
//   - Limit (int64): The limit of products to return.
//   - Offset (int64): The offset of products to return.
type ListProductsByCompanyIDPayload struct {
	CompanyID string `json:"company_id"`
	Limit    int64  `json:"limit"`
	Offset   int64  `json:"offset"`
}

// ListProductsResponse struct.
// This struct is used to represent a list products response.
//
// Attributes:
//   - Products ([]*Product): The list of products.
//   - TotalCount (int64): The total count of products.
type ListProductsResponse struct {
	Products   []*Product `json:"products"`
	TotalCount int64      `json:"total_count"`
}

// ProductDeletedEventPayload struct.
// This struct is used to represent a product delete payload.
//
// Attributes:
//   - ID (string): The product ID.
type ProductDeletedEventPayload struct {
	ID string `json:"id"`
}

type BaseModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CreatedAt time.Time          `bson:"createdAt" json:"created_at"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updated_at,omitempty"`
}
