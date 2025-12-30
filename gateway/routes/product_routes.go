package routes

import (
	consulDiscovery "common/discovery/consul"
	"gateway/controllers/product_controller"
	"gateway/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProductRoutes function.
// This function is used to define the product routes.
//
// Parameters:
//   - router (*gin.RouterGroup): The router group.
//   - registry (*consulDiscovery.ConsulRegistry): The Consul registry.
//   - logger (*logrus.Logger): The logger.
//
// Routes:
//   - POST /product: Create a product.
//   - PUT /product/:id: Update a product.
//   - GET /product: List products by vendor ID.
//   - DELETE /product/:id: Delete a product.
func ProductRoutes(router *gin.RouterGroup, registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) {
	productController := product_controller.NewProductController(registry, logger)

	productRouter := router.Group("/product")
	productRouter.Use(middlewares.AuthMiddleware(registry, logger))
	productRouter.POST("/", productController.CreateProduct)
	productRouter.PUT("/:id", productController.UpdateProduct)
	productRouter.GET("/", productController.ListProductsByCompanyID)
	productRouter.DELETE("/:id", productController.DeleteProduct)
}
