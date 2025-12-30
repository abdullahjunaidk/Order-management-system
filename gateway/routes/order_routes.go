package routes

import (
	consulDiscovery "common/discovery/consul"
	"gateway/controllers/order_controller"
	"gateway/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// OrderRoutes function.
// This function is used to define the order routes.
//
// Parameters:
//   - router (*gin.RouterGroup): The router group.
//   - registry (*consulDiscovery.ConsulRegistry): The Consul registry.
//   - logger (*logrus.Logger): The logger.
//
// Routes:
//   - POST /order/webhook: Handle Stripe webhook.
//   - POST /order: Create an order.
//   - GET /order/:id: Get an order by ID.
//   - DELETE /order/:id: Cancel an order.
func OrderRoutes(router *gin.RouterGroup, registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) {
	orderController := order_controller.NewOrderController(registry, logger)

	orderRouter := router.Group("/order")
	orderRouter.POST("/webhook", orderController.HandleStripeWebhook)
	orderRouter.Use(middlewares.AuthMiddleware(registry, logger))
	orderRouter.POST("", orderController.CreateOrder)
	orderRouter.GET("/:id", orderController.GetOrderByID)
	orderRouter.DELETE("/:id", orderController.CancelOrder)
}
