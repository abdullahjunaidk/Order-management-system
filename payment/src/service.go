package src

import (
	authSrc "auth/src"
	rabbitmqBroker "common/broker/rabbitmq"
	"common/helpers/env"
	"common/helpers/mailer"
	"context"
	"encoding/json"
	orderSrc "order/src"
	"os"
	"os/signal"
	productSrc "product/src"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v81"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// Project Domain
	PROJECT_DOMAIN = env.GetEnv("PROJECT_DOMAIN", "http://localhost:8080")

	// RabbitMQ Configuration.
	PRODUCT_CREATED_QUEUE         = env.GetEnv("PRODUCT_CREATED_QUEUE", "product.product.created")
	PRODUCT_UPDATED_QUEUE         = env.GetEnv("PRODUCT_UPDATED_QUEUE", "product.product.updated")
	PRODUCT_DELETED_QUEUE         = env.GetEnv("PRODUCT_DELETED_QUEUE", "product.product.deleted")
	ORDER_CREATED_QUEUE           = env.GetEnv("ORDER_CREATED_QUEUE", "order.order.created")
	ORDER_PENDING_CANCELLED_QUEUE = env.GetEnv("ORDER_PENDING_CANCELLED_QUEUE", "order.pending.cancelled")
	ORDER_PAID_SUCCESS_QUEUE      = env.GetEnv("ORDER_PAID_SUCCESS_QUEUE", "order.paid.success")
	ORDER_PAID_CANCELLED_QUEUE    = env.GetEnv("ORDER_PAID_CANCELLED_QUEUE", "order.paid.cancelled")

	// Mailpit Configuration.
	SMTP_USER     = env.GetEnv("SMTP_USER", "")
	SMTP_PASSWORD = env.GetEnv("SMTP_PASSWORD", "")
	SMTP_HOST     = env.GetEnv("SMTP_HOST", "localhost")
	SMTP_PORT     = env.GetEnvAsInt("SMTP_PORT", 1025)
	FROM_EMAIL    = env.GetEnv("FROM_EMAIL", "no-reply@example.com")
)

// PaymentService struct.
// This struct is used to define the payment service.
//
// Attributes:
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
//   - mailer (mailer.Mailer): The mailer.
//   - log (*logrus.Logger): The logger.
//   - authClient (*authSrc.AuthClient): The auth client.
//   - productClient (*productSrc.ProductClient): The product client.
//   - orderClient (*orderSrc.OrderClient): The order client.
//   - stripeAdapter (*StripeAdapter): The Stripe adapter.
type PaymentService struct {
	rabbitMQ      *rabbitmqBroker.RabbitMQAdapter
	mailer        mailer.Mailer
	log           *logrus.Logger
	authClient    *authSrc.AuthClient
	productClient *productSrc.ProductClient
	orderClient   *orderSrc.OrderClient
	stripeAdapter *StripeAdapter
}

// NewPaymentService function.
// This function is used to create a new payment service.
//
// Parameters:
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
//   - log (*logrus.Logger): The logger.
//   - productClient (*productSrc.ProductClient): The product client.
//   - orderClient (*orderSrc.OrderClient): The order client.
//   - stripeAdapter (*StripeAdapter): The Stripe adapter.
//
// Returns:
//   - *PaymentService: The payment service.
func NewPaymentService(rabbitMQ *rabbitmqBroker.RabbitMQAdapter, log *logrus.Logger, authClient *authSrc.AuthClient, productClient *productSrc.ProductClient, orderClient *orderSrc.OrderClient, stripeAdapter *StripeAdapter) *PaymentService {
	return &PaymentService{
		rabbitMQ:      rabbitMQ,
		mailer:        mailer.NewMailer(SMTP_USER, SMTP_PASSWORD, SMTP_HOST, SMTP_PORT, FROM_EMAIL, log),
		log:           log,
		authClient:    authClient,
		productClient: productClient,
		orderClient:   orderClient,
		stripeAdapter: stripeAdapter,
	}
}

// consumeMessages function.
// This function is used to consume messages from a specified queue.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - queueName (string): The queue name.
//   - handler (func([]byte)): The function to handle messages.
func (s *PaymentService) consumeMessages(ctx context.Context, queueName string, handler func([]byte)) {
	messages := make(chan []byte)

	go func() {
		if err := s.rabbitMQ.ConsumeMessages(ctx, queueName, messages); err != nil {
			s.log.WithFields(logrus.Fields{"queue": queueName, "error": err}).Error("Failed to Consume Messages!")
			panic(err)
		}
	}()

	go func() {
		for message := range messages {
			handler(message)
		}
	}()
}

// handleProductCreatedMessage function.
// This function is used to process product created messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *PaymentService) handleProductCreatedMessage(message []byte) {
	tracer := otel.Tracer("payment-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "paymentGRPCServer.consumeProductCreatedMessage")
	defer span.End()

	var payload ProductCreatedEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"product_id": payload.ID, "product_name": payload.Name, "vendor_id": payload.VendorID}).Info("Creating Stripe Product...")

	_, err := s.stripeAdapter.CreateProduct(ctx, payload.ID, payload.Name, payload.Description)
	if err != nil {
		s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID, "error": err}).Error("Failed to Create Stripe Product!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID}).Info("Creating Stripe Price...")

	stripePriceObj, err := s.stripeAdapter.CreatePrice(ctx, payload.ID, int64(payload.Price*100), "usd")
	if err != nil {
		s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID, "error": err}).Error("Failed to Create Stripe Price!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID}).Info("Setting Product Price ID...")

	if err := s.productClient.SetProductPriceID(ctx, payload.ID, payload.VendorID, stripePriceObj.ID); err != nil {
		s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID, "error": err}).Error("Failed to Set Product Price ID!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Product Price ID Set Successfully!")
	s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID}).Info("Product Price ID Set Successfully!")
}

// handleProductUpdatedMessage function.
// This function is used to process product updated messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *PaymentService) handleProductUpdatedMessage(message []byte) {
	tracer := otel.Tracer("payment-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "paymentGRPCServer.consumeProductUpdatedMessage")
	defer span.End()

	var payload ProductUpdatedEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"product_id": payload.ID, "product_name": payload.Name, "vendor_id": payload.VendorID}).Info("Updating Stripe Product...")

	_, err := s.stripeAdapter.UpdateProduct(ctx, payload.ID, payload.Name, payload.Description)
	if err != nil {
		s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID, "error": err}).Error("Failed to Update Stripe Product!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID}).Info("Updating Stripe Price...")

	stripePriceObj, err := s.stripeAdapter.UpdatePrice(ctx, payload.PriceID, int64(payload.Price*100), "usd")
	if err != nil {
		s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID, "error": err}).Error("Failed to Update Stripe Price!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"product_id": payload.ID, "vendor_id": payload.VendorID}).Info("Product Price Updated Successfully!")
	span.SetStatus(otlpcodes.Ok, "Product Price Updated Successfully!")

	s.log.WithFields(logrus.Fields{"price_id": stripePriceObj.ID, "product_id": payload.ID, "vendor_id": payload.VendorID}).Info("Stripe Product & Price Updated Successfully!")
}

// handleProductDeletedMessage function.
// This function is used to process product deleted messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *PaymentService) handleProductDeletedMessage(message []byte) {
	tracer := otel.Tracer("payment-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "paymentGRPCServer.consumeProductDeletedMessage")
	defer span.End()

	var payload ProductDeletedEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"product_id": payload.ID}).Info("Deleting Stripe Product...")

	if err := s.stripeAdapter.DeleteProduct(ctx, payload.ID); err != nil {
		s.log.WithFields(logrus.Fields{"product_id": payload.ID, "error": err}).Error("Failed to Delete Stripe Product!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"product_id": payload.ID}).Info("Stripe Product Deleted Successfully!")
}

// handleOrderCreatedMessage function.
// This function is used to process order created messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *PaymentService) handleOrderCreatedMessage(message []byte) {
	tracer := otel.Tracer("payment-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "paymentGRPCServer.consumeOrderCreatedMessage")
	defer span.End()

	var payload OrderCreatedEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID}).Info("Creating Payment Intent...")

	customer, err := s.authClient.GetUserByID(ctx, payload.CustomerID)
	if err != nil {
		s.log.WithFields(logrus.Fields{"customer_id": payload.CustomerID, "error": err}).Error("Failed to Get Customer!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	lineItems := make([]*stripe.PaymentLinkLineItemParams, len(payload.OrderItems))
	for i, item := range payload.OrderItems {
		lineItems[i] = &stripe.PaymentLinkLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(item.Quantity),
		}
	}

	paymentLink, err := s.stripeAdapter.GeneratePaymentLink(ctx, payload.OrderID, payload.CustomerID, lineItems)
	if err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Generate Payment Link!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	if err := s.orderClient.SetOrderPaymentLinkID(ctx, payload.OrderID, payload.CustomerID, paymentLink.ID); err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Set Order Payment Link ID!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID}).Info("Payment Link Generated Successfully!")

	data := map[string]interface{}{
		"CustomerUsername": customer.Username,
		"OrderID":          payload.OrderID,
		"PaymentLink":      paymentLink.URL,
	}

	if err := s.mailer.SendMailWithTemplate([]string{customer.Email}, "Order Confirmation", "templates/order.tmpl", data); err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Send Order Confirmation Email!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Order Confirmation Email Sent Successfully!")
	s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID}).Info("Order Confirmation Email Sent Successfully!")
}

// handleOrderPendingCancelledMessage function.
// This function is used to process order cancelled messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *PaymentService) handleOrderPendingCancelledMessage(message []byte) {
	tracer := otel.Tracer("payment-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "paymentGRPCServer.consumeOrderPendingCancelledMessage")
	defer span.End()

	var payload OrderCancelledEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID}).Info("Order Cancelled...")

	customer, err := s.authClient.GetUserByID(ctx, payload.CustomerID)
	if err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Get Customer!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	order, err := s.orderClient.GetOrderByIDAndCustomerID(ctx, payload.OrderID, payload.CustomerID)
	if err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Get Order!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	if order.PaymentLinkId != "" {
		if err := s.stripeAdapter.DeactivatePaymentLink(ctx, order.PaymentLinkId); err != nil {
			s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "payment_link_id": order.PaymentLinkId, "error": err}).Error("Failed to Deactivate Payment Link!")
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return
		}
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "payment_link_id": order.PaymentLinkId}).Info("Payment Link Deactivated Successfully!")
	}

	data := map[string]interface{}{
		"CustomerUsername": customer.Username,
		"OrderID":          payload.OrderID,
	}

	if err := s.mailer.SendMailWithTemplate([]string{customer.Email}, "Order Cancellation", "templates/order-cancelled.tmpl", data); err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Send Order Cancellation Email!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Order Cancelled Successfully!")
}

// handleOrderPaidSuccessMessage function.
// This function is used to process order paid messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *PaymentService) handleOrderPaidSuccessMessage(message []byte) {
	tracer := otel.Tracer("payment-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "paymentGRPCServer.consumeOrderPaidSuccessMessage")
	defer span.End()

	var payload OrderPaidSuccessEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID}).Info("Order Paid...")

	order, err := s.orderClient.GetOrderByIDAndCustomerID(ctx, payload.OrderID, payload.CustomerID)
	if err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Get Order!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	if order.PaymentLinkId != "" {
		if err := s.stripeAdapter.DeactivatePaymentLink(ctx, order.PaymentLinkId); err != nil {
			s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "payment_link_id": order.PaymentLinkId, "error": err}).Error("Failed to Deactivate Payment Link!")
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return
		}
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "payment_link_id": order.PaymentLinkId}).Info("Payment Link Deactivated Successfully!")
	}

	customer, err := s.authClient.GetUserByID(ctx, payload.CustomerID)
	if err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Get Customer!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	data := map[string]interface{}{
		"CustomerUsername": customer.Username,
		"OrderID":          payload.OrderID,
	}

	if err := s.mailer.SendMailWithTemplate([]string{customer.Email}, "Order Confirmation", "templates/order-paid.tmpl", data); err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Send Order Confirmation Email!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Order Paid Successfully!")
}

// handleOrderPaidCancelledMessage function.
// This function is used to process order cancelled messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *PaymentService) handleOrderPaidCancelledMessage(message []byte) {
	tracer := otel.Tracer("payment-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "paymentGRPCServer.consumeOrderPaidCancelledMessage")
	defer span.End()

	var payload OrderCancelledEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID}).Info("Order Cancelled...")

	customer, err := s.authClient.GetUserByID(ctx, payload.CustomerID)
	if err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Get Customer!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	order, err := s.orderClient.GetOrderByIDAndCustomerID(ctx, payload.OrderID, payload.CustomerID)
	if err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Get Order!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	if err := s.stripeAdapter.RefundPayment(ctx, order.PaymentId); err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "payment_id": order.PaymentId, "error": err}).Error("Failed to Refund Payment!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	data := map[string]interface{}{
		"CustomerUsername": customer.Username,
		"OrderID":          payload.OrderID,
	}

	if err := s.mailer.SendMailWithTemplate([]string{customer.Email}, "Order Cancellation", "templates/order-paid-cancelled.tmpl", data); err != nil {
		s.log.WithFields(logrus.Fields{"order_id": payload.OrderID, "customer_id": payload.CustomerID, "error": err}).Error("Failed to Send Order Cancellation Email!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Order Cancelled Successfully!")
}

// Run function.
// This function is used to run the payment service.
//
// Parameters:
//   - ctx (context.Context): The context.
//
// Returns:
//   - error: The error.
func (s *PaymentService) Run(ctx context.Context) error {
	go s.consumeMessages(ctx, PRODUCT_CREATED_QUEUE, s.handleProductCreatedMessage)
	go s.consumeMessages(ctx, PRODUCT_UPDATED_QUEUE, s.handleProductUpdatedMessage)
	go s.consumeMessages(ctx, PRODUCT_DELETED_QUEUE, s.handleProductDeletedMessage)
	go s.consumeMessages(ctx, ORDER_CREATED_QUEUE, s.handleOrderCreatedMessage)
	go s.consumeMessages(ctx, ORDER_PAID_CANCELLED_QUEUE, s.handleOrderPaidCancelledMessage)
	go s.consumeMessages(ctx, ORDER_PAID_SUCCESS_QUEUE, s.handleOrderPaidSuccessMessage)
	go s.consumeMessages(ctx, ORDER_PENDING_CANCELLED_QUEUE, s.handleOrderPendingCancelledMessage)

	<-ctx.Done()
	s.log.Info("Payment Service Stopped!")

	return nil
}

// GracefulShutdown function.
// This function is used to shutdown the payment service gracefully.
//
// Parameters:
//   - log (*logrus.Logger): The logger.
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
func GracefulShutdown(log *logrus.Logger, rabbitMQ *rabbitmqBroker.RabbitMQAdapter, productClient *productSrc.ProductClient) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan
	log.Warn("Shutting Down Gracefully...")

	if err := rabbitMQ.Close(); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close RabbitMQ Connection!")
	}
	log.Info("Disconnected from RabbitMQ!")

	if productClient != nil {
		productClient.Close()
		log.Info("Disconnected from Product Service!")
	}

	log.Info("Server Stopped!")
}
