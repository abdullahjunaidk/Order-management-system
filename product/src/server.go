package src

import (
	"common/helpers/logger"
	productProto "common/proto/product"
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

// productGRPCServer struct.
// This struct is used to implement the product service gRPC server.
//
// Attributes:
//   - productService (ProductService): The product service.
//   - log (*logrus.Logger): The logger.
type productGRPCServer struct {
	productProto.UnimplementedProductServiceServer
	productService ProductService
	log            *logrus.Logger
}

// ListenAndServeGRPC function.
// This function is used to listen and serve the product service gRPC server.
//
// Parameters:
//   - service (ProductService): The product service.
//   - port (int): The port.
//
// Returns:
//   - error: The error.
func ListenAndServeGRPC(service ProductService, port int) error {
	log := logger.NewLogger()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithFields(logrus.Fields{"port": port, "error": err}).Error("Failed to Listen!")
		return fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()

	productProto.RegisterProductServiceServer(server, &productGRPCServer{productService: service, log: log})
	grpc_health_v1.RegisterHealthServer(server, &healthServer{})
	reflection.Register(server)

	log.WithFields(logrus.Fields{"port": port}).Info("Starting gRPC Server...")
	if err := server.Serve(lis); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Serve gRPC Server!")
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// CreateProduct method.
// This method is used to create a new product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*productProto.ProductCreatePayload): The create product request.
//
// Returns:
//   - *productProto.Product: The created product.
//   - error: The error.
func (s *productGRPCServer) CreateProduct(ctx context.Context, req *productProto.ProductCreatePayload) (*productProto.Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productGRPCServer.CreateProduct")
	defer span.End()

	s.log.WithFields(logrus.Fields{"name": req.Name, "company_id": req.CompanyId}).Info("Creating New Product...")

	payload := ProductCreatePayload{
		CompanyID:   req.CompanyId,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       req.Price,
	}

	res, err := s.productService.CreateProduct(ctx, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"name": req.Name, "company_id": req.CompanyId, "error": err}).Error("Failed to Create Product!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"name": req.Name, "company_id": req.CompanyId}).Info("Product Created Successfully!")
	return &productProto.Product{
		Id:          res.ID.Hex(),
		CompanyId:   res.CompanyID,
		Name:        res.Name,
		Description: res.Description,
		Category:    res.Category,
		Price:       res.Price,
		PriceId:     res.PriceID,
		CreatedAt:   timestamppb.New(res.CreatedAt),
		UpdatedAt:   timestamppb.New(res.UpdatedAt),
	}, nil
}

// GetProductByIDAndCompanyID method.
// This method is used to get a product by ID and company ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*productProto.GetProductByIDAndCompanyIDPayload): The get product by ID and company ID request.
//
// Returns:
//   - *productProto.Product: The product.
//   - error: The error.
func (s *productGRPCServer) GetProductByIDAndCompanyID(ctx context.Context, req *productProto.GetProductByIDAndCompanyIDPayload) (*productProto.Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productGRPCServer.GetProductByIDAndCompanyID")
	defer span.End()

	s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId}).Info("Getting Product...")

	res, err := s.productService.GetProductByIDAndCompanyID(ctx, req.Id, req.CompanyId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId, "error": err}).Error("Failed to Get Product!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId}).Info("Product Found!")
	return &productProto.Product{
		Id:          res.ID.Hex(),
		CompanyId:   res.CompanyID,
		Name:        res.Name,
		Description: res.Description,
		Category:    res.Category,
		Price:       res.Price,
		PriceId:     res.PriceID,
		CreatedAt:   timestamppb.New(res.CreatedAt),
		UpdatedAt:   timestamppb.New(res.UpdatedAt),
	}, nil
}

// UpdateProduct method.
// This method is used to update a product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*productProto.ProductUpdatePayload): The update product request.
//
// Returns:
//   - *productProto.Product: The updated product.
//   - error: The error.
func (s *productGRPCServer) UpdateProduct(ctx context.Context, req *productProto.ProductUpdatePayload) (*productProto.Product, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productGRPCServer.UpdateProduct")
	defer span.End()

	s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId}).Info("Updating Product...")

	payload := ProductUpdatePayload{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       req.Price,
	}

	res, err := s.productService.UpdateProduct(ctx, req.Id, req.CompanyId, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId, "error": err}).Error("Failed to Update Product!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId}).Info("Product Updated Successfully!")
	return &productProto.Product{
		Id:          res.ID.Hex(),
		CompanyId:   res.CompanyID,
		Name:        res.Name,
		Description: res.Description,
		Category:    res.Category,
		Price:       res.Price,
		PriceId:     res.PriceID,
		CreatedAt:   timestamppb.New(res.CreatedAt),
		UpdatedAt:   timestamppb.New(res.UpdatedAt),
	}, nil
}

// ListProductsByCompanyID method.
// This method is used to list products by company ID with pagination.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*productProto.ListProductsByCompanyIDPayload): The list products by company ID request.
//
// Returns:
//   - *productProto.ListProductsResponse: The list products response.
//   - error: The error.
func (s *productGRPCServer) ListProductsByCompanyID(ctx context.Context, req *productProto.ListProductsByCompanyIDPayload) (*productProto.ListProductsResponse, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productGRPCServer.ListProductsByCompanyID")
	defer span.End()

	s.log.WithFields(logrus.Fields{"company_id": req.CompanyId, "limit": req.Limit, "offset": req.Offset}).Info("Listing Products...")

	res, err := s.productService.ListProductsByCompanyID(ctx, req.CompanyId, req.Limit, req.Offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"company_id": req.CompanyId, "limit": req.Limit, "offset": req.Offset, "error": err}).Error("Failed to List Products!")
		return nil, err
	}

	products := []*productProto.Product{}
	for _, product := range res.Products {
		products = append(products, &productProto.Product{
			Id:          product.ID.Hex(),
			CompanyId:   product.CompanyID,
			Name:        product.Name,
			Description: product.Description,
			Category:    product.Category,
			Price:       product.Price,
			PriceId:     product.PriceID,
			CreatedAt:   timestamppb.New(product.CreatedAt),
			UpdatedAt:   timestamppb.New(product.UpdatedAt),
		})
	}

	s.log.WithFields(logrus.Fields{"company_id": req.CompanyId, "limit": req.Limit, "offset": req.Offset}).Info("Products Listed Successfully!")
	return &productProto.ListProductsResponse{
		Products:   products,
		TotalCount: res.TotalCount,
	}, nil
}

// DeleteProduct method.
// This method is used to delete a product.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*productProto.DeleteProductPayload): The delete product request.
//
// Returns:
//   - *emptypb.Empty: The empty response.
//   - error: The error.
func (s *productGRPCServer) DeleteProduct(ctx context.Context, req *productProto.DeleteProductPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productGRPCServer.DeleteProduct")
	defer span.End()

	s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId}).Info("Deleting Product...")

	err := s.productService.DeleteProduct(ctx, req.Id, req.CompanyId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId, "error": err}).Error("Failed to Delete Product!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId}).Info("Product Deleted Successfully!")
	return &emptypb.Empty{}, nil
}

// SetProductPriceID method.
// This method is used to set product price ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*productProto.SetProductPriceIDPayload): The set product price ID request.
//
// Returns:
//   - *emptypb.Empty: The empty response.
//   - error: The error.
func (s *productGRPCServer) SetProductPriceID(ctx context.Context, req *productProto.SetProductPriceIDPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("product-service")
	ctx, span := tracer.Start(ctx, "productGRPCServer.SetProductPriceID")
	defer span.End()

	s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId, "price_id": req.PriceId}).Info("Setting Product Price ID...")

	err := s.productService.SetProductPriceID(ctx, req.Id, req.CompanyId, req.PriceId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId, "price_id": req.PriceId, "error": err}).Error("Failed to Set Product Price ID!")
		return nil, err
	}

	s.log.WithFields(logrus.Fields{"id": req.Id, "company_id": req.CompanyId, "price_id": req.PriceId}).Info("Product Price ID Set Successfully!")
	return &emptypb.Empty{}, nil
}
