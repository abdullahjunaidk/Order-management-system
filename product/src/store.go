package src

import (
	"context"

	mongoDatabase "common/database/mongo"

	otlpcodes "go.opentelemetry.io/otel/codes"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
)

var (
	// MongoDB collection name.
	COLLECTION_PRODUCTS = "products"
)

// ProductStore interface.
// This interface is used to define the product store methods.
//
// Methods:
//   - CreateProduct(ctx context.Context, product *Product) (string, error): This method is used to register a new product.
//   - GetProductByIDAndVendorID(ctx context.Context, productID string, vendorID string) (*Product, error): This method is used to get a product by ID and vendor ID.
//   - UpdateProduct(ctx context.Context, productID string, vendorID string, product *Product) error: This method is used to update a product.
//   - ListProductsByBrandID(ctx context.Context, brandID string, limit int64, offset int64) (*ListProductsResponse, error): This method is used to list products by brand ID with pagination.
//   - DeleteProduct(ctx context.Context, productID string, brandID string) error: This method is used to delete a product by ID and brand ID.
//   - SetProductPriceID(ctx context.Context, productID string, brandID string, priceID string) error: This method is used to set product price ID.
type ProductStore interface {
	CreateProduct(ctx context.Context, product *Product) (string, error)
	GetProductByID(ctx context.Context, productID string) (*Product, error)
	GetProductByIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Product, error)
	UpdateProduct(ctx context.Context, productID string, companyID string, product *Product) error
	ListProductsByCompanyID(ctx context.Context, companyID string, limit int64, offset int64) (*ListProductsResponse, error)
	DeleteProduct(ctx context.Context, productID string, companyID string) error
	SetProductPriceID(ctx context.Context, productID string, companyID string, priceID string) error
}

// productStore struct.
// This struct is used to implement the ProductStore interface.
//
// Attributes:
//   - productsCollection (*mongo.Collection): The products collection.
type productStore struct {
	productsCollection *mongo.Collection
}

// NewProductStore function.
// This function is used to create a new product store.
//
// Parameters:
//   - adapter (mongoDatabase.MongoDBAdapter): The MongoDB adapter.
//
// Returns:
//   - ProductStore: The product store.
func NewProductStore(adapter mongoDatabase.MongoDBAdapter) ProductStore {
	return &productStore{
		productsCollection: adapter.Collection(COLLECTION_PRODUCTS),
	}
}

// GetProductByID method.
// This method is used to get a product by ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//
// Returns:
//   - *Product: The product.
//   - error: An error if occurred.
func (store *productStore) GetProductByID(ctx context.Context, productID string) (*Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productStore.GetProductByID")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	filter := map[string]interface{}{
		"_id": objectID,
	}

	product := &Product{}
	err = store.productsCollection.FindOne(ctx, filter).Decode(product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	return product, nil
}

// CreateProduct method.
// This method is used to register a new product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - product (*Product): The product.
//
// Returns:
//   - string: The product ID.
//   - error: An error if occurred.
func (store *productStore) CreateProduct(ctx context.Context, product *Product) (string, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productStore.CreateProduct")
	defer span.End()

	res, err := store.productsCollection.InsertOne(ctx, product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return "", err
	}

	insertedID := res.InsertedID.(primitive.ObjectID).Hex()
	return insertedID, nil
}

// GetProductByIDAndVendorID method.
// This method is used to get a product by ID and vendor ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - *Product: The product.
//   - error: An error if occurred.
func (store *productStore) GetProductByIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productStore.GetProductByIDAndCompanyID")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	filter := map[string]interface{}{
		"_id":        objectID,
		"company_id": companyID,
	}

	product := &Product{}
	err = store.productsCollection.FindOne(ctx, filter).Decode(product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	return product, nil
}

// UpdateProduct method.
// This method is used to update a product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//   - product (*Product): The product.
//
// Returns:
//   - error: An error if occurred.
func (store *productStore) UpdateProduct(ctx context.Context, productID string, companyID string, product *Product) error {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productStore.UpdateProduct")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	filter := map[string]interface{}{
		"_id":        objectID,
		"company_id": companyID,
	}

	update := map[string]interface{}{
		"$set": product,
	}

	_, err = store.productsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}
	return nil
}

// ListProductsByCompanyID method.
// This method is used to list products by company ID with pagination.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - companyID (string): The company ID.
//   - limit (int64): The limit of products to return.
//   - offset (int64): The offset of products to return.
//
// Returns:
//   - *ListProductsResponse: The list products response.
//   - error: An error if occurred.
func (store *productStore) ListProductsByCompanyID(ctx context.Context, companyID string, limit int64, offset int64) (*ListProductsResponse, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productStore.ListProductsByCompanyID")
	defer span.End()

	filter := map[string]interface{}{"company_id": companyID}
	findOptions := options.Find().SetLimit(limit).SetSkip(offset)

	cursor, err := store.productsCollection.Find(ctx, filter, findOptions)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*Product
	if err := cursor.All(ctx, &products); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	totalCount, err := store.productsCollection.CountDocuments(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	response := &ListProductsResponse{
		Products:   products,
		TotalCount: totalCount,
	}

	return response, nil
}

// DeleteProduct method.
// This method is used to delete a product by ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - error: An error if occurred.
func (store *productStore) DeleteProduct(ctx context.Context, productID string, companyID string) error {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productStore.DeleteProduct")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	filter := map[string]interface{}{
		"_id":        objectID,
		"company_id": companyID,
	}

	_, err = store.productsCollection.DeleteOne(ctx, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Product Deleted Successfully!")
	return nil
}

// SetProductPriceID method.
// This method is used to set product price ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//   - priceID (string): The price ID.
//
// Returns:
//   - error: An error if occurred.
func (store *productStore) SetProductPriceID(ctx context.Context, productID string, companyID string, priceID string) error {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productStore.SetProductPriceID")
	defer span.End()

	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	filter := map[string]interface{}{
		"_id":        objectID,
		"company_id": companyID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{"price_id": priceID},
	}

	_, err = store.productsCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Product Price ID set successfully!")
	return nil
}
