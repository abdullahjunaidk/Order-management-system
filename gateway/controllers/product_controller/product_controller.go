package product_controller

import (
	authSrc "auth/src"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"errors"
	"gateway/controllers/auth_controller"
	"gateway/controllers/inventory_controller"
	"gateway/models/auth_models"
	"gateway/models/product_models"
	"net/http"

	authProto "common/proto/auth"
	inventorySrc "inventory/src"
	productSrc "product/src"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// Microservices Configuration
	PRODUCT_SERVICE_NAME = env.GetEnv("PRODUCT_SERVICE_NAME", "product-service")
)

// ProductController struct.
// This struct is used to represent an Product controller.
//
// Attributes:
//   - authClient (*authSrc.AuthClient): The Auth client.
//   - productClient (*productSrc.ProductClient): The Product client.
//   - logger (*logrus.Logger): The logger.
type ProductController struct {
	authClient      *authSrc.AuthClient
	productClient   *productSrc.ProductClient
	inventoryClient *inventorySrc.InventoryClient
	logger          *logrus.Logger
}

// NewProductController function.
// This function is used to create a new Product controller.
//
// Parameters:
//   - registry (*consulDiscovery.ConsulRegistry): The Consul registry.
//   - logger (*logrus.Logger): The logger.
//
// Returns:
//   - *ProductController: The Product controller.
func NewProductController(registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) *ProductController {
	authServiceAddress, err := registry.GetServiceAddress(auth_controller.AUTH_SERVICE_NAME)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Auth Service Address!")
	}

	authServiceClient, err := authSrc.NewAuthClient(authServiceAddress, "gateway-service")
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Auth Service Client!")
	}

	productServiceAddress, err := registry.GetServiceAddress(PRODUCT_SERVICE_NAME)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Product Service Address!")
	}

	productServiceClient, err := productSrc.NewProductClient(productServiceAddress, "gateway-service")
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Product Service Client!")
	}

	inventoryServiceAddress, err := registry.GetServiceAddress(inventory_controller.INVENTORY_SERVICE_NAME)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Inventory Service Address!")
	}

	inventoryServiceClient, err := inventorySrc.NewInventoryClient(inventoryServiceAddress, "gateway-service")
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Inventory Service Client!")
	}

	return &ProductController{
		authClient:      authServiceClient,
		productClient:   productServiceClient,
		inventoryClient: inventoryServiceClient,
		logger:          logger,
	}
}

// CreateProduct		godoc
// @Summary				Create a new product
// @Description			Create a new product on the Platform
// @Tags				Product
// @Accept				json
// @Produce				json
// @Security			BearerAuth
// @Param				Authorization header string true "Bearer Token"
// @Param				payload body models.ProductCreatePayload true "Product Create Payload"
// @Success				201 {object} models.ProductCreateSuccessResponse "Product Created Successfully!"
// @Failure				400 {object} models.ProductCreateErrorResponse "Failed to Create Product!"
// @Failure				401 {object} models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure				500 {object} models.ProductCreateErrorResponse "Internal Server Error!"
// @Router				/product [post]
func (pc *ProductController) CreateProduct(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.CreateProduct")
	defer span.End()

	var req product_models.ProductCreatePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Product!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, product_models.ProductCreateErrorResponse{
			Message: "Failed to Create Product!",
			Error:   err.Error(),
		})
		return
	}

	if err := auth_controller.Validate.Struct(req); err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Product!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, product_models.ProductCreateErrorResponse{
			Message: "Failed to Create Product!",
			Error:   err.Error(),
		})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("Failed to Create Product!")

		span.RecordError(errors.New("User not found"))
		span.SetStatus(otlpcodes.Error, "User not found")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "User not found",
		})
		return
	}

	authUser := user.(*authProto.User)
	if !authUser.IsSuperAdmin {

		pc.logger.Error("Failed to Create Product!")

		span.RecordError(errors.New("User doesn't have access to this company"))
		span.SetStatus(otlpcodes.Error, "User doesn't have access to this company")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "User doesn't have access to this company",
		})
		return
	}

	res, err := pc.productClient.CreateProduct(ctx, req.CompanyID, req.Name, req.Description, req.Category, req.Price)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Product!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, product_models.ProductCreateErrorResponse{
			Message: "Failed to Create Product!",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, product_models.ProductCreateSuccessResponse{
		Message: "Product Created Successfully!",
		Product: product_models.Product{
			ID:          res.Id,
			CompanyID:   res.CompanyId,
			Name:        res.Name,
			Description: res.Description,
			Category:    res.Category,
			Price:       res.Price,
			NetWeight:   res.NetWeight,
			CreatedAt:   res.CreatedAt.AsTime(),
			UpdatedAt:   res.UpdatedAt.AsTime(),
		},
	})

}

// UpdateProduct		godoc
// @Summary				Update a product
// @Description			Update a product on the Platform
// @Tags				Product
// @Accept				json
// @Produce				json
// @Security			BearerAuth
// @Param				Authorization header string true "Bearer Token"
// @Param				id path string true "Product ID"
// @Param				payload body product_models.ProductUpdatePayload true "Product Update Payload"
// @Success				200 {object} product_models.ProductUpdateSuccessResponse "Product Updated Successfully!"
// @Failure				400 {object} product_models.ProductUpdateErrorResponse "Failed to Update Product!"
// @Failure				401 {object} auth_models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure				500 {object} product_models.ProductUpdateErrorResponse "Internal Server Error!"
// @Router				/product/{id} [put]
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.UpdateProduct")
	defer span.End()

	productID := c.Param("id")

	var req product_models.ProductUpdatePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Update Product!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, product_models.ProductUpdateErrorResponse{
			Message: "Failed to Update Product!",
			Error:   err.Error(),
		})
		return
	}

	if err := auth_controller.Validate.Struct(req); err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Update Product!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, product_models.ProductUpdateErrorResponse{
			Message: "Failed to Update Product!",
			Error:   err.Error(),
		})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("Failed to Update Product!")

		span.RecordError(errors.New("user not found"))
		span.SetStatus(otlpcodes.Error, "user not found")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "user not found",
		})
		return
	}

	authUser := user.(*authProto.User)
	if !authUser.IsSuperAdmin {
		pc.logger.Error("Failed to Update Product!")

		span.RecordError(errors.New("user is not a vendor"))
		span.SetStatus(otlpcodes.Error, "user is not a vendor")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only vendors can update products",
		})
		return
	}

	res, err := pc.productClient.UpdateProduct(ctx, productID, authUser.Id, req.Name, req.Description, req.Category, req.Price)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Update Product!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, product_models.ProductUpdateErrorResponse{
			Message: "Failed to Update Product!",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, product_models.ProductUpdateSuccessResponse{
		Message: "Product Updated Successfully!",
		Product: product_models.Product{
			ID:          res.Id,
			CompanyID:   res.CompanyId,
			Name:        res.Name,
			Description: res.Description,
			Category:    res.Category,
			Price:       res.Price,
			NetWeight:   res.NetWeight,
			CreatedAt:   res.CreatedAt.AsTime(),
			UpdatedAt:   res.UpdatedAt.AsTime(),
		},
	})
}

// ListProductsByVendorID		godoc
// @Summary					List products by vendor ID
// @Description				List products for a vendor on the Platform with pagination
// @Tags					Product
// @Accept					json
// @Produce					json
// @Security				BearerAuth
// @Param					Authorization header string true "Bearer Token"
// @Param					limit query int false "Limit" default(10)
// @Param					offset query int false "Offset" default(0)
// @Success					200 {object} product_models.ListProductsByCompanyIDSuccessResponse "Products Listed Successfully!"
// @Failure					400 {object} product_models.ListProductsByCompanyIDErrorResponse "Failed to List Products!"
// @Failure					401 {object} auth_models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure					500 {object} product_models.ListProductsByCompanyIDErrorResponse "Internal Server Error!"
// @Router					/product [get]
func (pc *ProductController) ListProductsByCompanyID(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.ListProductsByCompanyID")
	defer span.End()

	var req product_models.ListProductsByCompanyIDPayload
	if err := c.ShouldBindQuery(&req); err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to List Products!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, product_models.ListProductsByCompanyIDErrorResponse{
			Message: "Failed to List Products!",
			Error:   err.Error(),
		})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("Failed to List Products!")

		span.RecordError(errors.New("user not found"))
		span.SetStatus(otlpcodes.Error, "user not found")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "user not found",
		})
		return
	}

	authUser := user.(*authProto.User)
	if !authUser.IsSuperAdmin {
		pc.logger.Error("Failed to List Products!")

		span.RecordError(errors.New("user is not a vendor"))
		span.SetStatus(otlpcodes.Error, "user is not a vendor")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only vendors can list products",
		})
		return
	}

	limit := req.Limit
	offset := req.Offset

	res, err := pc.productClient.ListProductsByCompanyID(ctx, req.CompanyID, limit, offset)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to List Products!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, product_models.ListProductsByCompanyIDErrorResponse{
			Message: "Failed to List Products!",
			Error:   err.Error(),
		})
		return
	}

	products := []product_models.Product{}
	for _, product := range res.Products {
		inventoryRes, err := pc.inventoryClient.GetInventoryByProductIDAndCompanyID(ctx, product.Id, req.CompanyID)
		if err != nil {
			pc.logger.WithFields(logrus.Fields{"error": err, "product_id": product.Id, "company_id": req.CompanyID}).Error("Failed to Get Inventory for Product!")
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())

			products = append(products, product_models.Product{
				ID:          product.Id,
				CompanyID:   product.CompanyId,
				Name:        product.Name,
				Description: product.Description,
				Category:    product.Category,
				Price:       product.Price,
				CreatedAt:   product.CreatedAt.AsTime(),
				UpdatedAt:   product.UpdatedAt.AsTime(),
				Inventory:   nil,
			})
			continue
		}
		products = append(products, product_models.Product{
			ID:          product.Id,
			CompanyID:   product.CompanyId,
			Name:        product.Name,
			Description: product.Description,
			Category:    product.Category,
			Price:       product.Price,
			CreatedAt:   product.CreatedAt.AsTime(),
			UpdatedAt:   product.UpdatedAt.AsTime(),
			Inventory: &product_models.ProductInventory{
				AvailableQuantity: inventoryRes.AvailableQuantity,
				ThresholdQuantity: inventoryRes.ThresholdQuantity,
				CreatedAt:         inventoryRes.CreatedAt.AsTime(),
				UpdatedAt:         inventoryRes.UpdatedAt.AsTime(),
			},
		})
	}

	c.JSON(http.StatusOK, product_models.ListProductsByCompanyIDSuccessResponse{
		Message:    "Products Listed Successfully!",
		Products:   products,
		TotalCount: res.TotalCount,
	})
}

// DeleteProduct		godoc
// @Summary				Delete a product
// @Description			Delete a product on the Platform
// @Tags				Product
// @Accept				json
// @Produce				json
// @Security			BearerAuth
// @Param				Authorization header string true "Bearer Token"
// @Param				id path string true "Product ID"
// @Success				200 {object} product_models.DeleteProductSuccessResponse "Product Deleted Successfully!"
// @Failure				400 {object} product_models.DeleteProductErrorResponse "Failed to Delete Product!"
// @Failure				401 {object} auth_models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure				500 {object} product_models.DeleteProductErrorResponse "Internal Server Error!"
// @Router				/product/{id} [delete]
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.DeleteProduct")
	defer span.End()

	productID := c.Param("id")

	user, exists := c.Get("user")
	if !exists {
		pc.logger.Error("Failed to Delete Product!")

		span.RecordError(errors.New("user not found"))
		span.SetStatus(otlpcodes.Error, "user not found")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "user not found",
		})
		return
	}

	authUser := user.(*authProto.User)
	if !authUser.IsSuperAdmin {
		pc.logger.Error("Failed to Delete Product!")

		span.RecordError(errors.New("user is not a admin"))
		span.SetStatus(otlpcodes.Error, "user is not a admin")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only admins can delete products",
		})
		return
	}

	if productID == "" {
		pc.logger.Error("Failed to Delete Product!")

		span.RecordError(errors.New("product id is required"))
		span.SetStatus(otlpcodes.Error, "product id is required")

		c.JSON(http.StatusBadRequest, product_models.DeleteProductErrorResponse{
			Message: "Failed to Delete Product!",
			Error:   "product id is required",
		})
		return
	}

	err := pc.productClient.DeleteProduct(ctx, productID, authUser.Id)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Delete Product!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, product_models.DeleteProductErrorResponse{
			Message: "Failed to Delete Product!",
			Error:   err.Error(),
		})
		return
	}

	err = pc.inventoryClient.DeleteInventoryByProductIDAndCompanyID(ctx, productID, authUser.Id)
	if err != nil {
		pc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Delete Inventory for Product!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
	} else {
		pc.logger.WithFields(logrus.Fields{"product_id": productID}).Info("Inventory Deleted Successfully for Product!")
	}

	c.JSON(http.StatusOK, product_models.DeleteProductSuccessResponse{
		Message: "Product Deleted Successfully!",
	})
}
