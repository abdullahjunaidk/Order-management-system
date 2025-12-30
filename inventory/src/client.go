package src

import (
	"common/helpers/logger"
	inventoryProto "common/proto/inventory"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InventoryClient struct.
// This struct is used to create a new client connection to the inventory service.
//
// Attributes:
//   - conn (*grpc.ClientConn): The client connection to the inventory service.
//   - service (inventoryProto.InventoryServiceClient): The inventory service client.
//   - tracerName (string): The tracer name.
//   - log (*logrus.Logger): The logger.
type InventoryClient struct {
	conn       *grpc.ClientConn
	service    inventoryProto.InventoryServiceClient
	tracerName string
	log        *logrus.Logger
}

// NewInventoryClient function.
// This function is used to create a new client connection to the inventory service.
//
// Parameters:
//   - address (string): The address of the inventory service.
//
// Returns:
//   - *InventoryClient: The inventory client.
//   - error: The error.
func NewInventoryClient(address string, tracerName string) (*InventoryClient, error) {
	log := logger.NewLogger()

	clientOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, clientOpts...)
	if err != nil {
		log.WithFields(logrus.Fields{"address": address, "error": err}).Error("Failed to Create Client Connection!")
		return nil, fmt.Errorf("failed to create client connection: %w", err)
	}

	service := inventoryProto.NewInventoryServiceClient(conn)

	return &InventoryClient{
		conn:       conn,
		service:    service,
		tracerName: tracerName,
		log:        log,
	}, nil
}

// Close function.
// This function is used to close the client connection to the inventory service.
func (c *InventoryClient) Close() {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			c.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close Connection!")
		} else {
			c.log.Info("Connection Closed!")
		}
	}
}

// CreateInventory method.
// This method is used to create a new inventory.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//   - availableQuantity (int64): The available quantity.
//   - thresholdQuantity (int64): The threshold quantity.
//
// Returns:
//   - *inventoryProto.Inventory: The created inventory.
//   - error: The error.
func (c *InventoryClient) CreateInventory(ctx context.Context, productID string, companyID string, availableQuantity int64, thresholdQuantity int64) (*inventoryProto.Inventory, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "InventoryClient.CreateInventory")
	defer span.End()

	req := &inventoryProto.CreateInventoryPayload{
		ProductId:         productID,
		CompanyId:          companyID,
		AvailableQuantity: availableQuantity,
		ThresholdQuantity: thresholdQuantity,
	}

	res, err := c.service.CreateInventory(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"product_id": productID, "company_id": companyID, "error": err}).Error("Failed to Create Inventory!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"product_id": productID, "company_id": companyID}).Info("Inventory Created Successfully!")
	return res, nil
}

// GetInventoryByProductIDAndVendorID method.
// This method is used to get inventory by product ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - *inventoryProto.Inventory: The inventory.
//   - error: The error.
func (c *InventoryClient) GetInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) (*inventoryProto.Inventory, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "InventoryClient.GetInventoryByProductIDAndCompanyID")
	defer span.End()

	req := &inventoryProto.GetInventoryByProductIDAndCompanyIDPayload{
		ProductId: productID,
		CompanyId: companyID,
	}

	res, err := c.service.GetInventoryByProductIDAndCompanyID(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"product_id": productID, "company_id": companyID, "error": err}).Error("Failed to Get Inventory!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"product_id": productID, "company_id": companyID}).Info("Inventory Fetched Successfully!")
	return res, nil
}

// DeleteInventoryByProductIDAndVendorID method.
// This method is used to delete inventory by product ID and vendor ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//
// Returns:
//   - error: The error.
func (c *InventoryClient) DeleteInventoryByProductIDAndCompanyID(ctx context.Context, productID string, companyID string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "InventoryClient.DeleteInventoryByProductIDAndCompanyID")
	defer span.End()

	req := &inventoryProto.DeleteInventoryByProductIDAndCompanyIDPayload{
		ProductId: productID,
		CompanyId: companyID,
	}

	_, err := c.service.DeleteInventoryByProductIDAndCompanyID(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"product_id": productID, "company_id": companyID, "error": err}).Error("Failed to Delete Inventory!")
		return err
	}

	c.log.WithFields(logrus.Fields{"product_id": productID, "company_id": companyID}).Info("Inventory Deleted Successfully!")
	return nil
}

// UpdateInventory method.
// This method is used to update inventory.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - productID (string): The product ID.
//   - companyID (string): The company ID.
//   - availableQuantity (int64): The available quantity.
//   - thresholdQuantity (int64): The threshold quantity.
//
// Returns:
//   - *inventoryProto.Inventory: The updated inventory.
//   - error: The error.
func (c *InventoryClient) UpdateInventory(ctx context.Context, productID string, companyID string, availableQuantity int64, thresholdQuantity int64) (*inventoryProto.Inventory, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "InventoryClient.UpdateInventory")
	defer span.End()

	req := &inventoryProto.UpdateInventoryPayload{
		ProductId:         productID,
		CompanyId:         companyID,
		AvailableQuantity: availableQuantity,
		ThresholdQuantity: thresholdQuantity,
	}

	res, err := c.service.UpdateInventory(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.log.WithFields(logrus.Fields{"product_id": productID, "company_id": companyID, "error": err}).Error("Failed to Update Inventory!")
		return nil, err
	}

	c.log.WithFields(logrus.Fields{"product_id": productID, "company_id": companyID}).Info("Inventory Updated Successfully!")
	return res, nil
}
