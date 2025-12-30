package src

import (
	"common/helpers/logger"
	inventoryProto "common/proto/inventory"
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// inventoryGRPCServer struct.
// This struct is used to implement the inventory service gRPC server.
//
// Attributes:
//   - inventoryService (InventoryService): The inventory service.
//   - log (*logrus.Logger): The logger.
type inventoryGRPCServer struct {
	inventoryProto.UnimplementedInventoryServiceServer
	inventoryService InventoryService
	log              *logrus.Logger
}

// ListenAndServeGRPC function.
// This function is used to listen and serve the inventory service gRPC server.
//
// Parameters:
//   - service (InventoryService): The inventory service.
//   - port (int): The port.
//
// Returns:
//   - error: The error.
func ListenAndServeGRPC(service InventoryService, port int) error {
	log := logger.NewLogger()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithFields(logrus.Fields{"port": port, "error": err}).Error("Failed to Listen!")
		return fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()

	inventoryProto.RegisterInventoryServiceServer(server, &inventoryGRPCServer{inventoryService: service, log: log})
	grpc_health_v1.RegisterHealthServer(server, &healthServer{})
	reflection.Register(server)

	log.WithFields(logrus.Fields{"port": port}).Info("Starting gRPC Server...")
	if err := server.Serve(lis); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Serve gRPC Server!")
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// CreateInventory method.
// This method is used to create a new inventory.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*inventoryProto.CreateInventoryPayload): The create inventory request.
//
// Returns:
//   - *inventoryProto.Inventory: The created inventory.
//   - error: The error.
func (s *inventoryGRPCServer) CreateInventory(ctx context.Context, req *inventoryProto.CreateInventoryPayload) (*inventoryProto.Inventory, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryGRPCServer.CreateInventory")
	defer span.End()

	s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "vendor_id": req.CompanyId}).Info("Creating New Inventory...")

	payload := CreateInventoryPayload{
		ProductID:         req.ProductId,
		CompanyID:         req.CompanyId,
		AvailableQuantity: req.AvailableQuantity,
		ThresholdQuantity: req.ThresholdQuantity,
	}

	res, err := s.inventoryService.CreateInventory(ctx, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "vendor_id": req.CompanyId, "error": err}).Error("Failed to Create Inventory!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "vendor_id": req.CompanyId}).Info("Inventory Created Successfully!")
	return &inventoryProto.Inventory{
		Id:                res.ID.Hex(),
		ProductId:         res.ProductID,
		CompanyId:         res.CompanyID,
		AvailableQuantity: res.AvailableQuantity,
		ThresholdQuantity: res.ThresholdQuantity,
		CreatedAt:         timestamppb.New(res.CreatedAt),
		UpdatedAt:         timestamppb.New(res.UpdatedAt),
	}, nil
}

// GetInventoryByProductIDAndCompanyID method.
// This method is used to get inventory by product ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*inventoryProto.GetInventoryByProductIDAndCompanyIDPayload): The get inventory by product ID and company ID request.
//
// Returns:
//   - *inventoryProto.Inventory: The inventory.
//   - error: The error.
func (s *inventoryGRPCServer) GetInventoryByProductIDAndCompanyID(ctx context.Context, req *inventoryProto.GetInventoryByProductIDAndCompanyIDPayload) (*inventoryProto.Inventory, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryGRPCServer.GetInventoryByProductIDAndCompanyID")
	defer span.End()

	s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId}).Info("Getting Inventory...")

	res, err := s.inventoryService.GetInventoryByProductIDAndCompanyID(ctx, req.ProductId, req.CompanyId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId, "error": err}).Error("Failed to Get Inventory!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId}).Info("Inventory Fetched Successfully!")
	return &inventoryProto.Inventory{
		Id:                res.ID.Hex(),
		ProductId:         res.ProductID,
		CompanyId:         res.CompanyID,
		AvailableQuantity: res.AvailableQuantity,
		ThresholdQuantity: res.ThresholdQuantity,
		CreatedAt:         timestamppb.New(res.CreatedAt),
		UpdatedAt:         timestamppb.New(res.UpdatedAt),
	}, nil
}

// DeleteInventoryByProductIDAndCompanyID method.
// This method is used to delete inventory by product ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*inventoryProto.DeleteInventoryByProductIDAndCompanyIDPayload): The delete inventory by product ID and company ID request.
//
// Returns:
//   - *emptypb.Empty: The empty response.
//   - error: The error.
func (s *inventoryGRPCServer) DeleteInventoryByProductIDAndCompanyID(ctx context.Context, req *inventoryProto.DeleteInventoryByProductIDAndCompanyIDPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryGRPCServer.DeleteInventoryByProductIDAndCompanyID")
	defer span.End()

	s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId}).Info("Deleting Inventory...")

	err := s.inventoryService.DeleteInventoryByProductIDAndCompanyID(ctx, req.ProductId, req.CompanyId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId, "error": err}).Error("Failed to Delete Inventory!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId}).Info("Inventory Deleted Successfully!")
	return &emptypb.Empty{}, nil
}

// UpdateInventory method.
// This method is used to update inventory.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*inventoryProto.UpdateInventoryPayload): The update inventory request.
//
// Returns:
//   - *inventoryProto.Inventory: The updated inventory.
//   - error: The error.
func (s *inventoryGRPCServer) UpdateInventory(ctx context.Context, req *inventoryProto.UpdateInventoryPayload) (*inventoryProto.Inventory, error) {
	tracer := otel.Tracer("inventory-service")
	ctx, span := tracer.Start(ctx, "inventoryGRPCServer.UpdateInventory")
	defer span.End()

	s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId}).Info("Updating Inventory...")

	payload := UpdateInventoryPayload{
		ProductID:         req.ProductId,
		CompanyID:         req.CompanyId,
		AvailableQuantity: req.AvailableQuantity,
		ThresholdQuantity: req.ThresholdQuantity,
	}

	res, err := s.inventoryService.UpdateInventory(ctx, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId, "error": err}).Error("Failed to Update Inventory!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"product_id": req.ProductId, "company_id": req.CompanyId}).Info("Inventory Updated Successfully!")
	return &inventoryProto.Inventory{
		Id:                res.ID.Hex(),
		ProductId:         res.ProductID,
		CompanyId:         res.CompanyID,
		AvailableQuantity: res.AvailableQuantity,
		ThresholdQuantity: res.ThresholdQuantity,
		CreatedAt:         timestamppb.New(res.CreatedAt),
		UpdatedAt:         timestamppb.New(res.UpdatedAt),
	}, nil
}
