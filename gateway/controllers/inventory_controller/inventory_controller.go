package inventory_controller

import (
	authSrc "auth/src"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"errors"
	"gateway/controllers/auth_controller"
	"gateway/models/auth_models"
	"gateway/models/inventory_models"
	"net/http"

	authProto "common/proto/auth"
	inventorySrc "inventory/src"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// Microservices Configuration
	INVENTORY_SERVICE_NAME = env.GetEnv("INVENTORY_SERVICE_NAME", "inventory-service")
)

// InventoryController struct.
// This struct is used to represent an Inventory controller.
//
// Attributes:
//   - authClient (*authSrc.AuthClient): The Auth client.
//   - inventoryClient (*inventorySrc.InventoryClient): The Inventory client.
//   - logger (*logrus.Logger): The logger.
type InventoryController struct {
	authClient      *authSrc.AuthClient
	inventoryClient *inventorySrc.InventoryClient
	logger          *logrus.Logger
}

// NewInventoryController function.
// This function is used to create a new Inventory controller.
//
// Parameters:
//   - registry (*consulDiscovery.ConsulRegistry): The Consul registry.
//   - logger (*logrus.Logger): The logger.
//
// Returns:
//   - *InventoryController: The Inventory controller.
func NewInventoryController(registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) *InventoryController {
	authServiceAddress, err := registry.GetServiceAddress(auth_controller.AUTH_SERVICE_NAME)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Auth Service Address!")
	}

	authServiceClient, err := authSrc.NewAuthClient(authServiceAddress, "gateway-service")
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Auth Service Client!")
	}

	inventoryServiceAddress, err := registry.GetServiceAddress(INVENTORY_SERVICE_NAME)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Inventory Service Address!")
	}

	inventoryServiceClient, err := inventorySrc.NewInventoryClient(inventoryServiceAddress, "gateway-service")
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Inventory Service Client!")
	}

	return &InventoryController{
		authClient:      authServiceClient,
		inventoryClient: inventoryServiceClient,
		logger:          logger,
	}
}

// CreateInventory		godoc
// @Summary				Create a new inventory
// @Description			Create a new inventory for a product by vendor
// @Tags				Inventory
// @Accept				json
// @Produce				json
// @Security			BearerAuth
// @Param				Authorization header string true "Bearer Token"
// @Param				id path string true "Product ID"
// @Param				payload body models.CreateInventoryPayload true "Inventory Create Payload"
// @Success				201 {object} models.CreateInventorySuccessResponse "Inventory Created Successfully!"
// @Failure				400 {object} models.CreateInventoryErrorResponse "Failed to Create Inventory!"
// @Failure				401 {object} models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure				500 {object} models.CreateInventoryErrorResponse "Internal Server Error!"
// @Router				/product/{id}/inventory [post]
func (ic *InventoryController) CreateInventory(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.CreateInventory")
	defer span.End()

	productID := c.Param("id")

	var req inventory_models.CreateInventoryPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ic.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Inventory!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, inventory_models.CreateInventoryErrorResponse{
			Message: "Failed to Create Inventory!",
			Error:   err.Error(),
		})
		return
	}

	if err := auth_controller.Validate.Struct(req); err != nil {
		ic.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Inventory!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, inventory_models.CreateInventoryErrorResponse{
			Message: "Failed to Create Inventory!",
			Error:   err.Error(),
		})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		ic.logger.Error("Unauthorized User!")

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
		ic.logger.Error("Unauthorized User Role!")

		span.RecordError(errors.New("user is not a vendor"))
		span.SetStatus(otlpcodes.Error, "user is not a vendor")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only vendors can create inventory",
		})
		return
	}

	res, err := ic.inventoryClient.CreateInventory(ctx, productID, authUser.Id, req.AvailableQuantity, req.ThresholdQuantity)
	if err != nil {
		ic.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Inventory!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, inventory_models.CreateInventoryErrorResponse{
			Message: "Failed to Create Inventory!",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, inventory_models.CreateInventorySuccessResponse{
		Message: "Inventory Created Successfully!",
		Inventory: inventory_models.Inventory{
			ID:                res.Id,
			ProductID:         res.ProductId,
			CompanyID:          res.CompanyId,
			AvailableQuantity: res.AvailableQuantity,
			ThresholdQuantity: res.ThresholdQuantity,
			CreatedAt:         res.CreatedAt.AsTime(),
			UpdatedAt:         res.UpdatedAt.AsTime(),
		},
	})
}

// GetInventoryByProductIDAndVendorID		godoc
// @Summary				Get inventory by product ID
// @Description			Get inventory for a product by vendor
// @Tags				Inventory
// @Accept				json
// @Produce				json
// @Security			BearerAuth
// @Param				Authorization header string true "Bearer Token"
// @Param				id path string true "Product ID"
// @Success				200 {object} inventory_models.GetInventoryByProductIDAndCompanyIDSuccessResponse "Inventory Fetched Successfully!"
// @Failure				400 {object} inventory_models.GetInventoryByProductIDAndCompanyIDErrorResponse "Failed to Get Inventory!"
// @Failure				401 {object} auth_models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure				500 {object} inventory_models.GetInventoryByProductIDAndCompanyIDErrorResponse "Internal Server Error!"
// @Router				/product/{id}/inventory [get]
func (ic *InventoryController) GetInventoryByProductIDAndCompanyID(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GetInventoryByProductIDAndCompanyID")
	defer span.End()

	productID := c.Param("id")

	user, exists := c.Get("user")
	if !exists {
		ic.logger.Error("Unauthorized User!")

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
		ic.logger.Error("Unauthorized User Role!")

		span.RecordError(errors.New("user is not a vendor"))
		span.SetStatus(otlpcodes.Error, "user is not a vendor")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only vendors can get inventory",
		})
		return
	}

	res, err := ic.inventoryClient.GetInventoryByProductIDAndCompanyID(ctx, productID, authUser.Id)
	if err != nil {
		ic.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Get Inventory!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, inventory_models.GetInventoryByProductIDAndCompanyIDErrorResponse{
			Message: "Failed to Get Inventory!",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, inventory_models.GetInventoryByProductIDAndCompanyIDSuccessResponse{
		Message: "Inventory Fetched Successfully!",
		Inventory: inventory_models.Inventory{
			ID:                res.Id,
			ProductID:         res.ProductId,
			CompanyID:          res.CompanyId,
			AvailableQuantity: res.AvailableQuantity,
			ThresholdQuantity: res.ThresholdQuantity,
			CreatedAt:         res.CreatedAt.AsTime(),
			UpdatedAt:         res.UpdatedAt.AsTime(),
		},
	})
}

// DeleteInventoryByProductIDAndVendorID		godoc
// @Summary				Delete inventory by product ID
// @Description			Delete inventory for a product by vendor
// @Tags				Inventory
// @Accept				json
// @Produce				json
// @Security			BearerAuth
// @Param				Authorization header string true "Bearer Token"
// @Param				id path string true "Product ID"
// @Success				200 {object} inventory_models.DeleteInventoryByProductIDAndCompanyIDSuccessResponse "Inventory Deleted Successfully!"
// @Failure				400 {object} inventory_models.DeleteInventoryByProductIDAndCompanyIDErrorResponse "Failed to Delete Inventory!"
// @Failure				401 {object} auth_models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure				500 {object} inventory_models.DeleteInventoryByProductIDAndCompanyIDErrorResponse "Internal Server Error!"
// @Router				/product/{id}/inventory [delete]
func (ic *InventoryController) DeleteInventoryByProductIDAndCompanyID(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.DeleteInventoryByProductIDAndCompanyID")
	defer span.End()

	productID := c.Param("id")

	user, exists := c.Get("user")
	if !exists {
		ic.logger.Error("Unauthorized User!")

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
		ic.logger.Error("Unauthorized User Role!")

		span.RecordError(errors.New("user is not a vendor"))
		span.SetStatus(otlpcodes.Error, "user is not a vendor")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only vendors can delete inventory",
		})
		return
	}

	err := ic.inventoryClient.DeleteInventoryByProductIDAndCompanyID(ctx, productID, authUser.Id)
	if err != nil {
		ic.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Delete Inventory!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, inventory_models.DeleteInventoryByProductIDAndCompanyIDErrorResponse{
			Message: "Failed to Delete Inventory!",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, inventory_models.DeleteInventoryByProductIDAndCompanyIDSuccessResponse{
		Message: "Inventory Deleted Successfully!",
	})
}

// UpdateInventory		godoc
// @Summary				Update inventory
// @Description			Update inventory for a product by vendor
// @Tags				Inventory
// @Accept				json
// @Produce				json
// @Security			BearerAuth
// @Param				Authorization header string true "Bearer Token"
// @Param				id path string true "Product ID"
// @Param				payload body inventory_models.UpdateInventoryPayload true "Inventory Update Payload"
// @Success				200 {object} inventory_models.UpdateInventorySuccessResponse "Inventory Updated Successfully!"
// @Failure				400 {object} inventory_models.UpdateInventoryErrorResponse "Failed to Update Inventory!"
// @Failure				401 {object} auth_models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure				500 {object} inventory_models.UpdateInventoryErrorResponse "Internal Server Error!"
// @Router				/product/{id}/inventory [put]
func (ic *InventoryController) UpdateInventory(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.UpdateInventory")
	defer span.End()

	productID := c.Param("id")

	var req inventory_models.UpdateInventoryPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ic.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Update Inventory!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, inventory_models.UpdateInventoryErrorResponse{
			Message: "Failed to Update Inventory!",
			Error:   err.Error(),
		})
		return
	}

	if err := auth_controller.Validate.Struct(req); err != nil {
		ic.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Update Inventory!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, inventory_models.UpdateInventoryErrorResponse{
			Message: "Failed to Update Inventory!",
			Error:   err.Error(),
		})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		ic.logger.Error("Unauthorized User!")

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
		ic.logger.Error("Unauthorized User Role!")

		span.RecordError(errors.New("user is not a vendor"))
		span.SetStatus(otlpcodes.Error, "user is not a vendor")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only vendors can update inventory",
		})
		return
	}

	res, err := ic.inventoryClient.UpdateInventory(ctx, productID, authUser.Id, req.AvailableQuantity, req.ThresholdQuantity)
	if err != nil {
		ic.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Update Inventory!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, inventory_models.UpdateInventoryErrorResponse{
			Message: "Failed to Update Inventory!",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, inventory_models.UpdateInventorySuccessResponse{
		Message: "Inventory Updated Successfully!",
		Inventory: inventory_models.Inventory{
			ID:                res.Id,
			ProductID:         res.ProductId,
			CompanyID:          res.CompanyId,
			AvailableQuantity: res.AvailableQuantity,
			ThresholdQuantity: res.ThresholdQuantity,
			CreatedAt:         res.CreatedAt.AsTime(),
			UpdatedAt:         res.UpdatedAt.AsTime(),
		},
	})
}
