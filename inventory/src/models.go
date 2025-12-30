package src

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Inventory struct.
// This struct is used to represent an Inventory in the database.
// It mirrors the protobuf Inventory message.
//
// Attributes:
//   - ID (primitive.ObjectID): The inventory ID.
//   - ProductID (string): The product ID.
//   - VendorID (string): The vendor ID.
//   - AvailableQuantity (int64): The available quantity.
//   - ThresholdQuantity (int64): The threshold quantity.
//   - CreatedAt (time.Time): The creation time.
//   - UpdatedAt (time.Time): The update time.
type Inventory struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ProductID         string             `bson:"product_id" json:"product_id"`
	CompanyID         string             `bson:"company_id" json:"company_id"`
	AvailableQuantity int64              `bson:"available_quantity" json:"available_quantity"`
	ThresholdQuantity int64              `bson:"threshold_quantity" json:"threshold_quantity"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreateInventoryPayload struct.
// This struct is used to represent a request payload for creating inventory.
// It mirrors the protobuf CreateInventoryPayload message.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
//   - AvailableQuantity (int64): The available quantity.
//   - ThresholdQuantity (int64): The threshold quantity.
type CreateInventoryPayload struct {
	ProductID         string `json:"product_id"`
	CompanyID         string `json:"company_id"`
	AvailableQuantity int64  `json:"available_quantity"`
	ThresholdQuantity int64  `json:"threshold_quantity"`
}

// GetInventoryByProductIDAndVendorIDPayload struct.
// This struct is used to represent a request payload for getting inventory by product and vendor IDs.
// It mirrors the protobuf GetInventoryByProductIDAndCompanyIDPayload message.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
type GetInventoryByProductIDAndCompanyIDPayload struct {
	ProductID string `json:"product_id"`
	CompanyID string `json:"company_id"`
}

// DeleteInventoryByProductIDAndCompanyIDPayload struct.
// This struct is used to represent a request payload for deleting inventory by product and company IDs.
// It mirrors the protobuf DeleteInventoryByProductIDAndCompanyIDPayload message.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
type DeleteInventoryByProductIDAndCompanyIDPayload struct {
	ProductID string `json:"product_id"`
	CompanyID string `json:"company_id"`
}

// UpdateInventoryPayload struct.
// This struct is used to represent a request payload for updating inventory.
// It mirrors the protobuf UpdateInventoryPayload message.
//
// Attributes:
//   - ProductID (string): The product ID.
//   - CompanyID (string): The company ID.
//   - AvailableQuantity (int64): The available quantity.
//   - ThresholdQuantity (int64): The threshold quantity.
type UpdateInventoryPayload struct {
	ProductID         string `json:"product_id"`
	CompanyID         string `json:"company_id"`
	AvailableQuantity int64  `json:"available_quantity"`
	ThresholdQuantity int64  `json:"threshold_quantity"`
}
