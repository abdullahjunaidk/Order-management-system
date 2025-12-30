package src

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

// InventoryService interface.
// This interface is used to define the inventory service methods.
//
// Methods:
//   - CreateInventory(ctx context.Context, payload CreateInventoryPayload) (*Inventory, error): Create a new inventory.
//   - GetInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Inventory, error): Get inventory by product ID and company ID.
//   - DeleteInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) error: Delete inventory by product ID and company ID.
//   - UpdateInventory(ctx context.Context, payload UpdateInventoryPayload) (*Inventory, error): Update inventory.
type InventoryService interface {
	CreateInventory(ctx context.Context, payload CreateInventoryPayload) (*Inventory, error)
	GetInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Inventory, error)
	DeleteInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) error
	UpdateInventory(ctx context.Context, payload UpdateInventoryPayload) (*Inventory, error)
}

// inventoryService struct.
// This struct is used to implement the InventoryService interface.
//
// Attributes:
//   - store (InventoryStore): The inventory store.
type inventoryService struct {
	store InventoryStore
}

// NewInventoryService function.
// This function is used to create a new inventory service.
//
// Parameters:
//   - store (InventoryStore): The inventory store.
//
// Returns:
//   - InventoryService: The inventory service.
func NewInventoryService(store InventoryStore) InventoryService {
	return &inventoryService{
		store: store,
	}
}

// CreateInventory method.
// This method is used to create a new inventory.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - payload (CreateInventoryPayload): The inventory payload.
//
// Returns:
//   - *Inventory: The created inventory.
//   - error: An error if occurred.
func (s *inventoryService) CreateInventory(ctx context.Context, payload CreateInventoryPayload) (*Inventory, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryService.CreateInventory")
	defer span.End()

	existingInventory, _ := s.store.GetInventoryByProductIDAndCompanyID(ctx, payload.ProductID, payload.CompanyID)
	if existingInventory != nil {
		span.RecordError(errors.New("inventory already exists for this product and company"))
		span.SetStatus(otlpcodes.Error, "inventory already exists")
		return nil, errors.New("inventory already exists for this product and company")
	}

	inventory := &Inventory{
		ID:                primitive.NewObjectID(),
		ProductID:         payload.ProductID,
		CompanyID:         payload.CompanyID,
		AvailableQuantity: payload.AvailableQuantity,
		ThresholdQuantity: payload.ThresholdQuantity,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	_, err := s.store.CreateInventory(ctx, inventory)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	createdInventory, err := s.store.GetInventoryByProductIDAndCompanyID(ctx, payload.ProductID, payload.CompanyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Inventory Created Successfully!")
	return createdInventory, nil
}

// GetInventoryByProductIDAndCompanyID method.
// This method is used to get inventory by product ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - *Inventory: The inventory.
//   - error: An error if occurred.
func (s *inventoryService) GetInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Inventory, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryService.GetInventoryByProductIDAndCompanyID")
	defer span.End()

	inventory, err := s.store.GetInventoryByProductIDAndCompanyID(ctx, productID, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Inventory Fetched Successfully!")
	return inventory, nil
}

// DeleteInventoryByProductIDAndVendorID method.
// This method is used to delete inventory by product ID and vendor ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - vendorID (string): The vendor ID.
//
// Returns:
//   - error: An error if occurred.
func (s *inventoryService) DeleteInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) error {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryService.DeleteInventoryByProductIDAndCompanyID")
	defer span.End()

	err := s.store.DeleteInventoryByProductIDAndCompanyID(ctx, productID, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Inventory Deleted Successfully!")
	return nil
}

// UpdateInventory method.
// This method is used to update inventory.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - payload (UpdateInventoryPayload): The inventory payload.
//
// Returns:
//   - *Inventory: The updated inventory.
//   - error: An error if occurred.
func (s *inventoryService) UpdateInventory(ctx context.Context, payload UpdateInventoryPayload) (*Inventory, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryService.UpdateInventory")
	defer span.End()

	inventory, err := s.store.GetInventoryByProductIDAndCompanyID(ctx, payload.ProductID, payload.CompanyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	inventory.AvailableQuantity = payload.AvailableQuantity
	inventory.ThresholdQuantity = payload.ThresholdQuantity
	inventory.UpdatedAt = time.Now()

	err = s.store.UpdateInventory(ctx, payload.ProductID, payload.CompanyID, inventory)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	updatedInventory, err := s.store.GetInventoryByProductIDAndCompanyID(ctx, payload.ProductID, payload.CompanyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Inventory Updated Successfully!")
	return updatedInventory, nil
}
