package order_controller

import (
	authSrc "auth/src"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"encoding/json"
	"errors"
	"gateway/controllers/auth_controller"
	"gateway/models/auth_models"
	"gateway/models/order_models"
	"net/http"

	authProto "common/proto/auth"
	orderProto "common/proto/order"
	orderSrc "order/src"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// Microservices Configuration
	ORDER_SERVICE_NAME = env.GetEnv("ORDER_SERVICE_NAME", "order-service")

	// Stipe Webhook Configuration
	STRIPE_WEBHOOK_SECRET = env.GetEnv("STRIPE_WEBHOOK_SECRET", "<STRIPE_WEBHOOK_SECRET>")
)

// OrderController struct.
// This struct is used to represent an Order controller.
//
// Attributes:
//   - authClient (*authSrc.AuthClient): The Auth client.
//   - orderClient (*orderSrc.OrderClient): The Order client.
//   - logger (*logrus.Logger): The logger.
type OrderController struct {
	authClient  *authSrc.AuthClient
	orderClient *orderSrc.OrderClient
	logger      *logrus.Logger
}

// NewOrderController function.
// This function is used to create a new Order controller.
//
// Parameters:
//   - registry (*consulDiscovery.ConsulRegistry): The Consul registry.
//   - logger (*logrus.Logger): The logger.
//
// Returns:
//   - *OrderController: The Order controller.
func NewOrderController(registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) *OrderController {
	authServiceAddress, err := registry.GetServiceAddress(auth_controller.AUTH_SERVICE_NAME)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Auth Service Address!")
	}

	authServiceClient, err := authSrc.NewAuthClient(authServiceAddress, "gateway-service")
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Auth Service Client!")
	}

	orderServiceAddress, err := registry.GetServiceAddress(ORDER_SERVICE_NAME)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Order Service Address!")
	}

	orderServiceClient, err := orderSrc.NewOrderClient(orderServiceAddress, "gateway-service")
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Order Service Client!")
	}

	return &OrderController{
		authClient:  authServiceClient,
		orderClient: orderServiceClient,
		logger:      logger,
	}
}

func (oc *OrderController) SampleOrder(c *gin.Context){
	tracer := otel.Tracer("gateway_service")

	ctx := c.Request.Context()
	ctx, span :=  tracer.Start(ctx, "gatewayHTTPServer.SampleOrder")
	defer span.End()

	var req order_models.OrderCreatePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Sample Order")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, order_models.OrderCreateErrorResponse{
			Message: "Failed to Create Sample Order",
			Error: err.Error(),
		})
		return
	}

	//Validate
	if err := auth_controller.Validate.Struct(req); err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Sample Order")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, order_models.OrderCreateErrorResponse{
			Message: "Failed to Create Sample Order",
			Error: err.Error(),
		})
		return 
	}

	//validate phone number
	
	


}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order for a customer
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param payload body models.OrderCreatePayload true "Order Create Payload"
// @Success 201 {object} models.OrderCreateSuccessResponse "Order Created Successfully!"
// @Failure 400 {object} models.OrderCreateErrorResponse "Failed to Create Order!"
// @Failure 401 {object} models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure 500 {object} models.OrderCreateErrorResponse "Internal Server Error!"
// @Router /order [post]
func (oc *OrderController) CreateOrder(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.CreateOrder")
	defer span.End()

	var req order_models.OrderCreatePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Order!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, order_models.OrderCreateErrorResponse{
			Message: "Failed to Create Order!",
			Error:   err.Error(),
		})
		return
	}

	if err := auth_controller.Validate.Struct(req); err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Order!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, order_models.OrderCreateErrorResponse{
			Message: "Failed to Create Order!",
			Error:   err.Error(),
		})
		return
	}

	// Get Authenticated User.
	user, exists := c.Get("user")
	if !exists {
		oc.logger.Error("Unauthorized User!")

		span.RecordError(errors.New("user not found"))
		span.SetStatus(otlpcodes.Error, "user not found")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "user not found",
		})
		return
	}

	// Check if User is a Customer.
	authUser := user.(*authProto.User)
	if !authUser.IsSuperAdmin {
		oc.logger.Error("Unauthorized User Role!")

		span.RecordError(errors.New("user is not a customer"))
		span.SetStatus(otlpcodes.Error, "user is not a customer")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only customers can create orders",
		})
		return
	}

	// Create Order.
	orderItems := []*orderProto.OrderItem{}
	for _, item := range req.Items {
		orderItems = append(orderItems, &orderProto.OrderItem{
			ProductId: item.ProductID,
			CompanyId:  item.CompanyID,
			Quantity:  item.Quantity,
		})
	}

	res, err := oc.orderClient.CreateOrder(ctx, orderItems, authUser.Id)
	if err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Create Order!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, order_models.OrderCreateErrorResponse{
			Message: "Failed to Create Order!",
			Error:   err.Error(),
		})
		return
	}

	items := []order_models.OrderItem{}
	for _, item := range res.Items {
		items = append(items, order_models.OrderItem{
			ProductID: item.ProductId,
			CompanyID:  item.CompanyId,
			Quantity:  item.Quantity,
		})
	}

	c.JSON(http.StatusCreated, order_models.OrderCreateSuccessResponse{
		Message: "Order Created Successfully!",
		Order: order_models.Order{
			ID:            res.OrderId,
			Items:         items,
			CustomerID:    res.CustomerId,
			Status:        res.Status,
			TotalPrice:    res.TotalPrice,
			PaymentLinkID: res.PaymentLinkId,
			CreatedAt:     res.CreatedAt.AsTime(),
			UpdatedAt:     res.UpdatedAt.AsTime(),
		},
	})
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel an order for a customer
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param id path string true "Order ID"
// @Success 200 {object} order_models.CancelOrderSuccessResponse "Order Cancelled Successfully!"
// @Failure 400 {object} order_models.CancelOrderErrorResponse "Failed to Cancel Order!"
// @Failure 401 {object} auth_models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure 500 {object} order_models.CancelOrderErrorResponse "Internal Server Error!"
// @Router /order/{id} [delete]
func (oc *OrderController) CancelOrder(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.CancelOrder")
	defer span.End()

	orderID := c.Param("id")

	user, exists := c.Get("user")
	if !exists {
		oc.logger.Error("Unauthorized User!")

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
		oc.logger.Error("Unauthorized User Role!")

		span.RecordError(errors.New("user is not a customer"))
		span.SetStatus(otlpcodes.Error, "user is not a customer")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only customers can cancel orders",
		})
		return
	}

	err := oc.orderClient.CancelOrder(ctx, orderID, authUser.Id)
	if err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Cancel Order!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, order_models.CancelOrderErrorResponse{
			Message: "Failed to Cancel Order!",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, order_models.CancelOrderSuccessResponse{
		Message: "Order Cancelled Successfully!",
	})
}

// GetOrderByID godoc
// @Summary Get order by ID
// @Description Get order by ID for a customer
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param Authorization header string true "Bearer Token"
// @Param id path string true "Order ID"
// @Success 200 {object} order_models.GetOrderByIDSuccessResponse "Order Retrieved Successfully!"
// @Failure 400 {object} order_models.GetOrderByIDErrorResponse "Failed to Get Order!"
// @Failure 401 {object} auth_models.UnauthorizedErrorResponse "Unauthorized!"
// @Failure 500 {object} order_models.GetOrderByIDErrorResponse "Internal Server Error!"
// @Router /order/{id} [get]
func (oc *OrderController) GetOrderByID(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GetOrderByID")
	defer span.End()

	orderID := c.Param("id")

	user, exists := c.Get("user")
	if !exists {
		oc.logger.Error("Unauthorized User!")

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
		oc.logger.Error("Unauthorized User Role!")

		span.RecordError(errors.New("user is not a customer"))
		span.SetStatus(otlpcodes.Error, "user is not a customer")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only customers can get orders",
		})
		return
	}

	res, err := oc.orderClient.GetOrderByIDAndCustomerID(ctx, orderID, authUser.Id)
	if err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Get Order!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, order_models.GetOrderByIDErrorResponse{
			Message: "Failed to Get Order!",
			Error:   err.Error(),
		})
		return
	}

	items := []order_models.OrderItem{}
	for _, item := range res.Items {
		items = append(items, order_models.OrderItem{
			ProductID: item.ProductId,
			CompanyID:  item.CompanyId,
			Quantity:  item.Quantity,
		})
	}

	c.JSON(http.StatusOK, order_models.GetOrderByIDSuccessResponse{
		Message: "Order Retrieved Successfully!",
		Order: order_models.Order{
			ID:            res.OrderId,
			Items:         items,
			CustomerID:    res.CustomerId,
			Status:        res.Status,
			TotalPrice:    res.TotalPrice,
			PaymentLinkID: res.PaymentLinkId,
			CreatedAt:     res.CreatedAt.AsTime(),
			UpdatedAt:     res.UpdatedAt.AsTime(),
		},
	})
}

// HandleStripeWebhook godoc
// @Summary Handle Stripe webhook
// @Description Handle Stripe webhook events
// @Tags Order
// @Accept json
// @Produce json
// @Success 200 "Webhook Handled Successfully!"
// @Failure 400 "Bad Request"
// @Failure 500 "Internal Server Error!"
// @Router /order/webhook [post]
func (oc *OrderController) HandleStripeWebhook(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.HandleStripeWebhook")
	defer span.End()

	payload, err := c.GetRawData()
	if err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Get Raw Data from Webhook Request!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEventWithOptions(payload, c.Request.Header.Get("Stripe-Signature"), STRIPE_WEBHOOK_SECRET, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Construct Event from Webhook Request!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			oc.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Unmarshal Stripe Event!")
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if session.PaymentStatus == "paid" {
			orderID := session.Metadata["order_id"]
			if orderID == "" {
				oc.logger.WithFields(logrus.Fields{"event_id": event.ID}).Error("Order ID not found in Payment Link metadata!")
				span.RecordError(errors.New("order ID not found in Payment Link metadata"))
				span.SetStatus(otlpcodes.Error, "order ID not found in Payment Link metadata")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			customerID := session.Metadata["customer_id"]
			if customerID == "" {
				oc.logger.WithFields(logrus.Fields{"event_id": event.ID, "order_id": orderID}).Error("Customer ID not found in Payment Link metadata!")
				span.RecordError(errors.New("customer ID not found in Payment Link metadata"))
				span.SetStatus(otlpcodes.Error, "customer ID not found in Payment Link metadata")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			paymentID := session.PaymentIntent.ID
			if paymentID == "" {
				oc.logger.WithFields(logrus.Fields{"event_id": event.ID, "order_id": orderID}).Error("Payment ID not found in checkout session!")
				span.RecordError(errors.New("payment ID not found in checkout session"))
				span.SetStatus(otlpcodes.Error, "payment ID not found in checkout session")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}

			err := oc.orderClient.SetOrderPaid(ctx, orderID, customerID, paymentID)
			if err != nil {
				oc.logger.WithFields(logrus.Fields{"event_id": event.ID, "order_id": orderID, "customer_id": customerID, "error": err}).Error("Failed to Set Order to Paid!")
				span.RecordError(err)
				span.SetStatus(otlpcodes.Error, err.Error())
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			oc.logger.WithFields(logrus.Fields{"event_id": event.ID, "order_id": orderID, "customer_id": customerID}).Info("Order Set to Paid Successfully via Webhook!")
			c.Status(http.StatusOK)
		}

	default:
		oc.logger.WithFields(logrus.Fields{"event_type": event.Type, "event_id": event.ID}).Info("Unhandled Stripe Webhook Event Type!")
		c.Status(http.StatusOK)
	}
}
