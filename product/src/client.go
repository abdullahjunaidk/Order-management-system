package src

import (
	"common/helpers/logger"
	productProto "common/proto/product"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ProductClient struct.
// This struct is used to create a new client connection to the product service.
//
// Attributes:
//   - conn (*grpc.ClientConn): The client connection to the product service.
//   - service (productProto.ProductServiceClient): The product service client.
//   - tracerName (string): The tracer name.
//   - log (*logrus.Logger): The logger.
type ProductClient struct {
	conn       *grpc.ClientConn
	service    productProto.ProductServiceClient
	tracerName string
	log        *logrus.Logger
}

// NewProductClient function.
// This function is used to create a new client connection to the product service.
//
// Parameters:
//   - address (string): The address of the product service.
//
// Returns:
//   - *ProductClient: The product client.
//   - error: The error.
func NewProductClient(address string, tracerName string) (*ProductClient, error) {
	log := logger.NewLogger()

	clientOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Create a new client connection to the product service.
	conn, err := grpc.NewClient(address, clientOpts...)
	if err != nil {
		log.WithFields(logrus.Fields{"address": address, "error": err}).Error("Failed to Create Client Connection!")
		return nil, fmt.Errorf("failed to create client connection: %w", err)
	}

	service := productProto.NewProductServiceClient(conn)

	return &ProductClient{
		conn:       conn,
		service:    service,
		tracerName: tracerName,
		log:        log,
	}, nil
}

// Close function.
// This function is used to close the client connection to the product service.
func (c *ProductClient) Close() {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			c.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close Connection!")
		} else {
			c.log.Info("Connection Closed!")
		}
	}
}

// CreateProduct method.
// This method is used to create a new product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - companyID (string): The company ID.
//   - name (string): The name.
//   - description (string): The description.
//   - category (string): The category.
//   - price (float32): The price.
//
// Returns:
//   - *productProto.Product: The created product.
//   - error: The error.
func (c *ProductClient) CreateProduct(ctx context.Context, companyId string, name string, description string, category string, price float32) (*productProto.Product, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "ProductClient.CreateProduct")
	defer span.End()

	req := &productProto.ProductCreatePayload{
		CompanyId:   companyId,
		Name:        name,
		Description: description,
		Category:    category,
		Price:       price,
	}

	res, err := c.service.CreateProduct(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"company_id": companyId, "name": name, "error": err}).Error("Failed to Create Product!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"company_id": companyId, "name": name}).Info("Product Created Successfully!")
	return res, nil
}

// GetProductByIDAndCompanyID method.
// This method is used to get a product by ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - id (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - *productProto.Product: The product.
//   - error: The error.
func (c *ProductClient) GetProductByIDAndCompanyID(ctx context.Context, productId string, companyId string) (*productProto.Product, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "ProductClient.GetProductByIDAndCompanyID")
	defer span.End()

	req := &productProto.GetProductByIDAndCompanyIDPayload{
		Id:        productId,
		CompanyId: companyId,
	}

	res, err := c.service.GetProductByIDAndCompanyID(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"id": productId, "company_id": companyId, "error": err}).Error("Failed to Get Product!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"id": productId, "company_id": companyId}).Info("Product Retrieved Successfully!")
	return res, nil
}

func (c *ProductClient) GetProductByID(ctx context.Context, productId string) (*productProto.Product, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "ProductClient.GetProductByID")
	defer span.End()

	req := &productProto.GetProductByIDPayload{
		Id: productId,
	}

	res, err := c.service.GetProductByID(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"id": productId, "error": err}).Error("Failed to Get Product!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"id": productId}).Info("Product Retrieved Successfully!")
	return res, nil
}

// UpdateProduct method.
// This method is used to update a product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - id (string): The product ID.
//   - companyID (string): The company ID.
//   - name (string): The name.
//   - description (string): The description.
//   - category (string): The category.
//   - price (float32): The price.
//
// Returns:
//   - *productProto.Product: The updated product.
//   - error: The error.
func (c *ProductClient) UpdateProduct(ctx context.Context, id string, companyId string, name string, description string, category string, price float32) (*productProto.Product, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "ProductClient.UpdateProduct")
	defer span.End()

	req := &productProto.ProductUpdatePayload{
		Id:          id,
		CompanyId:   companyId,
		Name:        name,
		Description: description,
		Category:    category,
		Price:       price,
	}

	res, err := c.service.UpdateProduct(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"id": id, "company_id": companyId, "name": name, "error": err}).Error("Failed to Update Product!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"id": id, "company_id": companyId, "name": name}).Info("Product Updated Successfully!")
	return res, nil
}

// ListProductsByCompanyID method.
// This method is used to list products by company ID with pagination.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - companyID (string): The company ID.
//   - limit (int64): The limit.
//   - offset (int64): The offset.
//
// Returns:
//   - *productProto.ListProductsResponse: The list products response.
//   - error: The error.
func (c *ProductClient) ListProductsByCompanyID(ctx context.Context, companyID string, limit int64, offset int64) (*productProto.ListProductsResponse, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "ProductClient.ListProductsByCompanyID")
	defer span.End()

	req := &productProto.ListProductsByCompanyIDPayload{
		CompanyId: companyID,
		Limit:     limit,
		Offset:    offset,
	}

	res, err := c.service.ListProductsByCompanyID(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"company_id": companyID, "limit": limit, "offset": offset, "error": err}).Error("Failed to List Products!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"company_id": companyID, "limit": limit, "offset": offset}).Info("Products Listed Successfully!")
	return res, nil
}

// DeleteProduct method.
// This method is used to delete a product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - error: The error.
func (c *ProductClient) DeleteProduct(ctx context.Context, productID string, companyID string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "ProductClient.DeleteProduct")
	defer span.End()

	req := &productProto.DeleteProductPayload{
		Id:        productID,
		CompanyId: companyID,
	}

	_, err := c.service.DeleteProduct(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"id": productID, "company_id": companyID, "error": err}).Error("Failed to Delete Product!")
		return err
	}

	c.log.WithFields(logrus.Fields{"id": productID, "company_id": companyID}).Info("Product Deleted Successfully!")
	return nil
}

// SetProductPriceID method.
// This method is used to set product price id.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//   - priceID (string): The price ID.
//
// Returns:
//   - error: The error.
func (c *ProductClient) SetProductPriceID(ctx context.Context, productID string, companyID string, priceID string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "ProductClient.SetProductPriceID")
	defer span.End()

	req := &productProto.SetProductPriceIDPayload{
		Id:        productID,
		CompanyId: companyID,
		PriceId:   priceID,
	}

	_, err := c.service.SetProductPriceID(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"id": productID, "company_id": companyID, "price_id": priceID, "error": err}).Error("Failed to Set Product Price ID!")
		return err
	}

	c.log.WithFields(logrus.Fields{"id": productID, "company_id": companyID, "price_id": priceID}).Info("Product Price ID Set Successfully!")
	return nil
}
