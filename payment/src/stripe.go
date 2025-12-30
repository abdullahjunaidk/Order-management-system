package src

import (
	"common/helpers/env"
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentlink"
	"github.com/stripe/stripe-go/v81/price"
	"github.com/stripe/stripe-go/v81/product"
	"github.com/stripe/stripe-go/v81/refund"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// Stripe Configuration
	STRIPE_SECRET_KEY = env.GetEnv("STRIPE_SECRET_KEY", "")
)

// StripeAdapter struct.
// This struct is used to interact with the Stripe API.
//
// Attributes:
//   - apiKey (string): The Stripe API key.
type StripeAdapter struct {
	apiKey string
	log    *logrus.Logger
}

// NewStripeAdapter function.
// This function creates a new Stripe adapter.
//
// Parameters:
//   - log (*logrus.Logger): The logger.
//
// Returns:
//   - *StripeAdapter: The Stripe adapter.
//   - error: An error if the Stripe API key is not found in the environment variables.
func NewStripeAdapter(log *logrus.Logger) (*StripeAdapter, error) {
	if STRIPE_SECRET_KEY == "" {
		return nil, fmt.Errorf("STRIPE_SECRET_KEY not found in environment variables")
	}

	stripe.Key = STRIPE_SECRET_KEY
	return &StripeAdapter{apiKey: STRIPE_SECRET_KEY, log: log}, nil
}

// CreateProduct function.
// This function creates a new product in Stripe.
//
// Parameters:
//   - productID (string): The ID of the product.
//   - productName (string): The name of the product.
//   - productDescription (string): The description of the product.
//
// Returns:
//   - *stripe.Product: The Stripe product.
//   - error: An error if the product creation fails.
func (s *StripeAdapter) CreateProduct(ctx context.Context, productID, productName, productDescription string) (*stripe.Product, error) {
	tracer := otel.Tracer("payment-service")
	ctx, span := tracer.Start(ctx, "StripeAdapter.CreateProduct")
	defer span.End()

	params := &stripe.ProductParams{
		ID:          stripe.String(productID),
		Name:        stripe.String(productName),
		Description: stripe.String(productDescription),
	}

	p, err := product.New(params)
	if err != nil {
		s.log.WithFields(logrus.Fields{"productID": productID, "error": err}).Error("Failed to Create Product!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	s.log.WithFields(logrus.Fields{"productID": productID}).Info("Product Created Successfully!")
	span.SetStatus(otlpcodes.Ok, "Product Created Successfully!")
	return p, nil
}

// CreatePrice function.
// This function creates a price for a product in Stripe.
//
// Parameters:
//   - productID (string): The ID of the product.
//   - unitAmount (int64): The unit amount for the price.
//   - currency (string): The currency for the price.
//
// Returns:
//   - *stripe.Price: The Stripe price.
//   - error: An error if the price creation fails.
func (s *StripeAdapter) CreatePrice(ctx context.Context, productID string, unitAmount int64, currency string) (*stripe.Price, error) {
	tracer := otel.Tracer("payment-service")
	ctx, span := tracer.Start(ctx, "StripeAdapter.CreatePrice")
	defer span.End()

	params := &stripe.PriceParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(unitAmount),
		Currency:   stripe.String(currency),
	}

	p, err := price.New(params)
	if err != nil {
		s.log.WithFields(logrus.Fields{"productID": productID, "error": err}).Error("Failed to Create Price!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to create price: %w", err)
	}

	s.log.WithFields(logrus.Fields{"productID": productID}).Info("Price Created Successfully!")
	span.SetStatus(otlpcodes.Ok, "Price Created Successfully!")
	return p, nil
}

// UpdateProduct function.
// This function updates an existing product in Stripe.
//
// Parameters:
//   - productID (string): The ID of the product to update.
//   - productName (string): The new name of the product.
//   - productDescription (string): The new description of the product.
//
// Returns:
//   - *stripe.Product: The updated Stripe product.
//   - error: An error if the product update fails.
func (s *StripeAdapter) UpdateProduct(ctx context.Context, productID, productName, productDescription string) (*stripe.Product, error) {
	tracer := otel.Tracer("payment-service")
	ctx, span := tracer.Start(ctx, "StripeAdapter.UpdateProduct")
	defer span.End()

	params := &stripe.ProductParams{
		Name:        stripe.String(productName),
		Description: stripe.String(productDescription),
	}

	p, err := product.Update(productID, params)
	if err != nil {
		s.log.WithFields(logrus.Fields{"productID": productID, "error": err}).Error("Failed to Update Product!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	s.log.WithFields(logrus.Fields{"productID": productID}).Info("Product Updated Successfully!")
	span.SetStatus(otlpcodes.Ok, "Product Updated Successfully!")
	return p, nil
}

// UpdatePrice function.
// This function updates an existing price in Stripe.
//
// Parameters:
//   - priceID (string): The ID of the price to update.
//   - unitAmount (int64): The new unit amount for the price.
//   - currency (string): The currency for the price.
//
// Returns:
//   - *stripe.Price: The updated Stripe price.
//   - error: An error if the price update fails.
func (s *StripeAdapter) UpdatePrice(ctx context.Context, priceID string, unitAmount int64, currency string) (*stripe.Price, error) {
	tracer := otel.Tracer("payment-service")
	ctx, span := tracer.Start(ctx, "StripeAdapter.UpdatePrice")
	defer span.End()

	params := &stripe.PriceParams{
		UnitAmount: stripe.Int64(unitAmount),
		Currency:   stripe.String(currency),
	}

	p, err := price.Update(priceID, params)
	if err != nil {
		s.log.WithFields(logrus.Fields{"priceID": priceID, "error": err}).Error("Failed to Update Price!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to update price: %w", err)
	}

	s.log.WithFields(logrus.Fields{"priceID": priceID}).Info("Price Updated Successfully!")
	span.SetStatus(otlpcodes.Ok, "Price Updated Successfully!")
	return p, nil
}

// DeleteProduct function.
// This function deletes an existing product in Stripe.
//
// Parameters:
//   - productID (string): The ID of the product to delete.
//
// Returns:
//   - error: An error if the product deletion fails.
func (s *StripeAdapter) DeleteProduct(ctx context.Context, productID string) error {
	tracer := otel.Tracer("payment-service")
	ctx, span := tracer.Start(ctx, "StripeAdapter.DeleteProduct")
	defer span.End()

	_, err := product.Del(productID, nil)
	if err != nil {
		s.log.WithFields(logrus.Fields{"productID": productID, "error": err}).Error("Failed to Delete Product!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return fmt.Errorf("failed to delete product: %w", err)
	}

	s.log.WithFields(logrus.Fields{"productID": productID}).Info("Product Deleted Successfully!")
	span.SetStatus(otlpcodes.Ok, "Product Deleted Successfully!")
	return nil
}

// GeneratePaymentLink function.
// This function generates a payment link for an order in Stripe.
//
// Parameters:
//   - lineItems ([]*stripe.PaymentLinkLineItemParams): The line items for the payment link.
//
// Returns:
//   - *stripe.PaymentLink: The URL of the payment link.
//   - error: An error if the payment link generation fails.
func (s *StripeAdapter) GeneratePaymentLink(ctx context.Context, orderID string, customerID string, lineItems []*stripe.PaymentLinkLineItemParams) (*stripe.PaymentLink, error) {
	tracer := otel.Tracer("payment-service")
	ctx, span := tracer.Start(ctx, "StripeAdapter.GeneratePaymentLink")
	defer span.End()

	params := &stripe.PaymentLinkParams{
		LineItems: lineItems,
		Metadata: map[string]string{
			"order_id":    orderID,
			"customer_id": customerID,
		},
	}

	result, err := paymentlink.New(params)
	if err != nil {
		s.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Generate Payment Link!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to generate payment link: %w", err)
	}

	s.log.Info("Payment Link Generated Successfully!")
	span.SetStatus(otlpcodes.Ok, "Payment Link Generated Successfully!")
	return result, nil
}

// DeactivatePaymentLink function.
// This function deletes an existing payment link in Stripe.
//
// Parameters:
//   - paymentLinkID (string): The ID of the payment link to delete.
//
// Returns:
//   - error: An error if the payment link deletion fails.
func (s *StripeAdapter) DeactivatePaymentLink(ctx context.Context, paymentLinkID string) error {
	tracer := otel.Tracer("payment-service")
	ctx, span := tracer.Start(ctx, "StripeAdapter.DeactivatePaymentLink")
	defer span.End()

	_, err := paymentlink.Update(paymentLinkID, &stripe.PaymentLinkParams{Active: stripe.Bool(false)})
	if err != nil {
		s.log.WithFields(logrus.Fields{"paymentLinkID": paymentLinkID, "error": err}).Error("Failed to Delete Payment Link!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return fmt.Errorf("failed to delete payment link: %w", err)
	}

	s.log.WithFields(logrus.Fields{"paymentLinkID": paymentLinkID}).Info("Payment Link Deleted Successfully!")
	span.SetStatus(otlpcodes.Ok, "Payment Link Deleted Successfully!")
	return nil
}

// RefundPayment function.
// This function refunds a payment in Stripe.
//
// Parameters:
//   - paymentID (string): The ID of the payment to refund.
//
// Returns:
//   - error: An error if the payment refund fails.
func (s *StripeAdapter) RefundPayment(ctx context.Context, paymentID string) error {
	tracer := otel.Tracer("payment-service")
	ctx, span := tracer.Start(ctx, "StripeAdapter.RefundPayment")
	defer span.End()

	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentID),
	}

	_, err := refund.New(params)
	if err != nil {
		s.log.WithFields(logrus.Fields{"paymentID": paymentID, "error": err}).Error("Failed to Refund Payment!")
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return fmt.Errorf("failed to refund payment: %w", err)
	}

	s.log.WithFields(logrus.Fields{"paymentID": paymentID}).Info("Payment Refunded Successfully!")
	span.SetStatus(otlpcodes.Ok, "Payment Refunded Successfully!")
	return nil
}
