package src

import (
	rabbitmqBroker "common/broker/rabbitmq"
	"common/helpers/env"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// RabbitMQ Queues.
	PRODUCT_CREATED_QUEUE = env.GetEnv("PRODUCT_CREATED_QUEUE", "product.product.created")
	PRODUCT_UPDATED_QUEUE = env.GetEnv("PRODUCT_UPDATED_QUEUE", "product.product.updated")
	PRODUCT_DELETED_QUEUE = env.GetEnv("PRODUCT_DELETED_QUEUE", "product.product.deleted")
)

// ProductService interface.
// This interface is used to define the product service methods.
//
// Methods:
//   - CreateProduct(ctx context.Context, payload ProductCreatePayload) (*Product, error): This method is used to register a new product.
//   - GetProductByIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Product, error): This method is used to get a product by ID and company ID.
//   - UpdateProduct(ctx context.Context, productID string, companyID string, payload ProductUpdatePayload) (*Product, error): This method is used to update a product.
//   - ListProductsByCompanyID(ctx context.Context, companyID string, limit int64, offset int64) (*ListProductsResponse, error): This method is used to list products by company ID with pagination.
//   - DeleteProduct(ctx context.Context, productID string, companyID string) error: This method is used to delete a product by ID and company ID.
//   - SetProductPriceID(ctx context.Context, productID string, companyID string, priceID string) error: This method is used to set product price ID.
type ProductService interface {
	CreateProduct(ctx context.Context, payload ProductCreatePayload) (*Product, error)
	GetProductByID(ctx context.Context, productID string) (*Product, error)
	GetProductByIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Product, error)
	UpdateProduct(ctx context.Context, productID string, companyID string, payload ProductUpdatePayload) (*Product, error)
	ListProductsByCompanyID(ctx context.Context, companyID string, limit int64, offset int64) (*ListProductsResponse, error)
	DeleteProduct(ctx context.Context, productID string, companyID string) error
	SetProductPriceID(ctx context.Context, productID string, companyID string, priceID string) error
}

// productService struct.
// This struct is used to implement the ProductService interface.
//
// Attributes:
//   - store (ProductStore): The product store.
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
type productService struct {
	store    ProductStore
	rabbitMQ *rabbitmqBroker.RabbitMQAdapter
}

// NewProductService function.
// This function is used to create a new product service.
//
// Parameters:
//   - store (ProductStore): The product store.
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
//
// Returns:
//   - ProductService: The product service.
func NewProductService(store ProductStore, rabbitMQ *rabbitmqBroker.RabbitMQAdapter) ProductService {
	return &productService{
		store:    store,
		rabbitMQ: rabbitMQ,
	}
}

// CreateProduct method.
// This method is used to register a new product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - payload (ProductCreatePayload): The product create payload.
//
// Returns:
//   - product (*Product): The product.
//   - error: An error if the operation failed.
func (s *productService) CreateProduct(ctx context.Context, payload ProductCreatePayload) (*Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productService.CreateProduct")
	defer span.End()

	product := &Product{
		BaseModel: BaseModel{
			ID:        primitive.NewObjectID(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		CompanyID:   payload.CompanyID,
		Name:        payload.Name,
		Description: payload.Description,
		Category:    payload.Category,
		Price:       payload.Price,
		PriceID:     "",
	}

	createdProductID, err := s.store.CreateProduct(ctx, product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	createdProduct, err := s.store.GetProductByIDAndCompanyID(ctx, createdProductID, payload.CompanyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	eventPayload := ProductCreatedEventPayload{
		ID:          createdProduct.ID.Hex(),
		CompanyID:   createdProduct.CompanyID,
		Name:        createdProduct.Name,
		Description: createdProduct.Description,
		Category:    createdProduct.Category,
		Price:       createdProduct.Price,
	}

	if err := s.rabbitMQ.PublishMessage(ctx, PRODUCT_CREATED_QUEUE, eventPayload); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Product Created Successfully!")
	return createdProduct, nil
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
func (s *productService) GetProductByID(ctx context.Context, productID string) (*Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productService.GetProductByID")
	defer span.End()

	product, err := s.store.GetProductByID(ctx, productID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Product Found Successfully!")
	return product, nil
}

// GetProductByIDAndCompanyID method.
// This method is used to get a product by ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - *Product: The product.
//   - error: An error if occurred.
func (s *productService) GetProductByIDAndCompanyID(ctx context.Context, productID string, companyID string) (*Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productService.GetProductByIDAndCompanyID")
	defer span.End()

	product, err := s.store.GetProductByIDAndCompanyID(ctx, productID, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Product Found Successfully!")
	return product, nil
}

// UpdateProduct method.
// This method is used to update a product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//   - payload (ProductUpdatePayload): The product update payload.
//
// Returns:
//   - *Product: The product.
//   - error: An error if occurred.
func (s *productService) UpdateProduct(ctx context.Context, productID string, companyID string, payload ProductUpdatePayload) (*Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productService.UpdateProduct")
	defer span.End()

	product, err := s.store.GetProductByIDAndCompanyID(ctx, productID, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	product.Name = payload.Name
	product.Description = payload.Description
	product.Category = payload.Category
	product.Price = payload.Price
	product.UpdatedAt = time.Now()

	err = s.store.UpdateProduct(ctx, productID, companyID, product)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	updatedProduct, err := s.store.GetProductByIDAndCompanyID(ctx, productID, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	eventPayload := ProductUpdatedEventPayload{
		ID:          updatedProduct.ID.Hex(),
		CompanyID:   updatedProduct.CompanyID,
		Name:        updatedProduct.Name,
		Description: updatedProduct.Description,
		Category:    updatedProduct.Category,
		Price:       updatedProduct.Price,
	}

	if err := s.rabbitMQ.PublishMessage(ctx, PRODUCT_UPDATED_QUEUE, eventPayload); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Product Updated Successfully!")
	return updatedProduct, nil
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
func (s *productService) ListProductsByCompanyID(ctx context.Context, companyID string, limit int64, offset int64) (*ListProductsResponse, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productService.ListProductsByCompanyID")
	defer span.End()

	response, err := s.store.ListProductsByCompanyID(ctx, companyID, limit, offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Products Listed Successfully!")
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
func (s *productService) DeleteProduct(ctx context.Context, productID string, companyID string) error {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productService.DeleteProduct")
	defer span.End()

	err := s.store.DeleteProduct(ctx, productID, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	eventPayload := ProductDeletedEventPayload{
		ID: productID,
	}

	if err := s.rabbitMQ.PublishMessage(ctx, PRODUCT_DELETED_QUEUE, eventPayload); err != nil {
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
func (s *productService) SetProductPriceID(ctx context.Context, productID string, companyID string, priceID string) error {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productService.SetProductPriceID")
	defer span.End()

	err := s.store.SetProductPriceID(ctx, productID, companyID, priceID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Product Price ID set successfully!")
	return nil
}
