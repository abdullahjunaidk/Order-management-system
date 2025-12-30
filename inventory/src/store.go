package src

import (
	"context"
	"log"

	mongoDatabase "common/database/mongo"

	otlpcodes "go.opentelemetry.io/otel/codes"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
)

var (
	// MongoDB collection name.
	COLLECTION_INVENTORY = "inventory"
)

// InventoryStore interface.
// This interface is used to define the inventory store methods.
//
// Methods:
//   - CreateInventory(ctx context.Context, inventory *Inventory) (string, error): Create a new inventory.
//   - GetInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Inventory, error): Get inventory by product ID and company ID.
//   - DeleteInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) error: Delete inventory by product ID and company ID.
//   - UpdateInventory(ctx context.Context, productID string, companyID string, inventory *Inventory) error: Update inventory by product ID and company ID.
type InventoryStore interface {
	CreateInventory(ctx context.Context, inventory *Inventory) (string, error)
	GetInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Inventory, error)
	DeleteInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) error
	UpdateInventory(ctx context.Context, productID string, companyID string, inventory *Inventory) error
}

// inventoryStore struct.
// This struct is used to implement the InventoryStore interface.
//
// Attributes:
//   - inventoryCollection (*mongo.Collection): The users collection.
type inventoryStore struct {
	inventoryCollection *mongo.Collection
}

// NewInventoryStore function.
// This function is used to create a new inventory store.
//
// Parameters:
//   - adapter (mongoDatabase.MongoDBAdapter): The MongoDB adapter.
//
// Returns:
//   - InventoryStore: The inventory store.
func NewInventoryStore(adapter mongoDatabase.MongoDBAdapter) InventoryStore {
	collection := adapter.Collection(COLLECTION_INVENTORY)

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "product_id", Value: 1}, {Key: "company_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Fatal(err)
	}

	return &inventoryStore{
		inventoryCollection: collection,
	}
}

// CreateInventory method.
// This method is used to create a new inventory.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - inventory (*Inventory): The inventory.
//
// Returns:
//   - string: The inventory ID.
//   - error: An error if occurred.
func (s *inventoryStore) CreateInventory(ctx context.Context, inventory *Inventory) (string, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryStore.CreateInventory")
	defer span.End()

	res, err := s.inventoryCollection.InsertOne(ctx, inventory)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return "", err
	}

	insertedID := res.InsertedID.(primitive.ObjectID).Hex()
	return insertedID, nil
}

// GetInventoryByProductIDAndVendorID method.
// This method is used to get inventory by product ID and vendor ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - *Inventory: The inventory.
//   - error: An error if occurred.
func (s *inventoryStore) GetInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Inventory, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryStore.GetInventoryByProductIDAndCompanyID")
	defer span.End()

	filter := map[string]interface{}{
		"product_id": productID,
		"company_id": companyID,
	}

	var inventory Inventory
	err := s.inventoryCollection.FindOne(ctx, filter).Decode(&inventory)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	return &inventory, nil
}

// DeleteInventoryByProductIDAndVendorID method.
// This method is used to delete inventory by product ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - error: An error if occurred.
func (s *inventoryStore) DeleteInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) error {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryStore.DeleteInventoryByProductIDAndCompanyID")
	defer span.End()

	filter := map[string]interface{}{
		"product_id": productID,
		"company_id": companyID,
	}

	_, err := s.inventoryCollection.DeleteOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	return nil
}

// UpdateInventory method.
// This method is used to update inventory.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//   - inventory (*Inventory): The inventory.
//
// Returns:
//   - error: An error if occurred.
func (s *inventoryStore) UpdateInventory(ctx context.Context, productID string, companyID string, inventory *Inventory) error {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryStore.UpdateInventory")
	defer span.End()

	filter := map[string]interface{}{
		"product_id": productID,
		"company_id": companyID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"available_quantity": inventory.AvailableQuantity,
			"threshold_quantity": inventory.ThresholdQuantity,
			"updated_at":         inventory.UpdatedAt,
		},
	}

	_, err := s.inventoryCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	return nil
}
