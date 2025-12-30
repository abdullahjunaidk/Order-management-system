package routes

import (
	consulDiscovery "common/discovery/consul"
	"gateway/controllers/inventory_controller"
	"gateway/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// InventoryRoutes function.
// This function is used to define the inventory routes.
//
// Parameters:
//   - router (*gin.RouterGroup): The router group.
//   - registry (*consulDiscovery.ConsulRegistry): The Consul registry.
//   - logger (*logrus.Logger): The logger.
//
// Routes:
//   - POST /product/:id/inventory: Create a new inventory.
//   - GET /product/:id/inventory: Get inventory by product ID.
//   - DELETE /product/:id/inventory: Delete inventory by product ID.
//   - PUT /product/:id/inventory: Update inventory by product ID.
func InventoryRoutes(router *gin.RouterGroup, registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) {
	inventoryController := inventory_controller.NewInventoryController(registry, logger)

	inventoryRouter := router.Group("/product/:id/inventory")
	inventoryRouter.Use(middlewares.AuthMiddleware(registry, logger))
	inventoryRouter.POST("", inventoryController.CreateInventory)
	inventoryRouter.GET("", inventoryController.GetInventoryByProductIDAndCompanyID)
	inventoryRouter.DELETE("", inventoryController.DeleteInventoryByProductIDAndCompanyID)
	inventoryRouter.PUT("", inventoryController.UpdateInventory)
}
