package src

import (
	"common/helpers/env"
	"common/helpers/mailer"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	rabbitmqBroker "common/broker/rabbitmq"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// Project Domain
	PROJECT_DOMAIN = env.GetEnv("PROJECT_DOMAIN", "http://localhost:8080")

	// RabbitMQ Configuration.
	USER_REGISTERED_QUEUE      = env.GetEnv("USER_REGISTERED_QUEUE", "auth.user.registered")
	USER_ACTIVATED_QUEUE       = env.GetEnv("USER_ACTIVATED_QUEUE", "auth.user.activated")
	USER_FORGOT_PASSWORD_QUEUE = env.GetEnv("USER_FORGOT_PASSWORD_QUEUE", "auth.user.forgotPassword")
	USER_PASSWORD_RESET_QUEUE  = env.GetEnv("USER_PASSWORD_RESET_QUEUE", "auth.user.passwordReset")

	// Mailpit Configuration.
	SMTP_USER     = env.GetEnv("SMTP_USER", "")
	SMTP_PASSWORD = env.GetEnv("SMTP_PASSWORD", "")
	SMTP_HOST     = env.GetEnv("SMTP_HOST", "localhost")
	SMTP_PORT     = env.GetEnvAsInt("SMTP_PORT", 1025)
	FROM_EMAIL    = env.GetEnv("FROM_EMAIL", "no-reply@example.com")
)

// MailerService struct.
// This struct is used to define the mailer service.
//
// Attributes:
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
//   - mailer (mailer.Mailer): The mailer.
//   - log (*logrus.Logger): The logger.
type MailerService struct {
	rabbitMQ *rabbitmqBroker.RabbitMQAdapter
	mailer   mailer.Mailer
	log      *logrus.Logger
}

// NewMailerService function.
// This function is used to create a new mailer service.
//
// Parameters:
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
//   - log (*logrus.Logger): The logger.
//
// Returns:
//   - *MailerService: The mailer service.
func NewMailerService(rabbitMQ *rabbitmqBroker.RabbitMQAdapter, log *logrus.Logger) *MailerService {
	return &MailerService{
		rabbitMQ: rabbitMQ,
		mailer:   mailer.NewMailer(SMTP_USER, SMTP_PASSWORD, SMTP_HOST, SMTP_PORT, FROM_EMAIL, log),
		log:      log,
	}
}

// consumeMessages function.
// This function is used to consume messages from a specified queue.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - queueName (string): The queue name.
//   - handler (func([]byte)): The function to handle messages.
func (s *MailerService) consumeMessages(ctx context.Context, queueName string, handler func([]byte)) {
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

// handleUserRegisteredMessage function.
// This function is used to process user registered messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *MailerService) handleUserRegisteredMessage(message []byte) {
	tracer := otel.Tracer("mailer-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "mailerGRPCServer.consumeUserRegisteredMessage")
	defer span.End()

	var payload UserRegisteredEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"email": payload.Email}).Info("Sending Activation Email...")

	data := map[string]interface{}{
		"Username":       payload.Username,
		"ActivationLink": fmt.Sprintf("%s/api/v1/auth/activate/%s", PROJECT_DOMAIN, payload.ActivationToken),
	}

	if err := s.mailer.SendMailWithTemplate([]string{payload.Email}, "Activate Your Account", "templates/activation.tmpl", data); err != nil {
		s.log.WithFields(logrus.Fields{"email": payload.Email, "error": err}).Error("Failed to Send Activation Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Activation Email Sent Successfully!")
	s.log.WithFields(logrus.Fields{"email": payload.Email}).Info("Activation Email Sent Successfully!")
}

// handleUserActivatedMessage function.
// This function is used to process user activated messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *MailerService) handleUserActivatedMessage(message []byte) {
	tracer := otel.Tracer("mailer-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "mailerGRPCServer.consumeUserActivatedMessage")
	defer span.End()

	var payload UserActivatedEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"email": payload.Email}).Info("Sending Welcome Email...")

	data := map[string]interface{}{
		"Username":  payload.Username,
		"LoginLink": fmt.Sprintf("%s/api/v1/auth/login", PROJECT_DOMAIN),
	}

	if err := s.mailer.SendMailWithTemplate([]string{payload.Email}, "Welcome to the Club!", "templates/welcome.tmpl", data); err != nil {
		s.log.WithFields(logrus.Fields{"email": payload.Email, "error": err}).Error("Failed to Send Welcome Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Welcome Email Sent Successfully!")
	s.log.WithFields(logrus.Fields{"email": payload.Email}).Info("Welcome Email Sent Successfully!")
}

// handleUserForgotPasswordMessage function.
// This function is used to process user forgot password messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *MailerService) handleUserForgotPasswordMessage(message []byte) {
	tracer := otel.Tracer("mailer-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "mailerGRPCServer.consumeUserForgotPasswordMessage")
	defer span.End()

	var payload UserForgotPasswordEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"email": payload.Email}).Info("Sending Forgot Password Email...")

	data := map[string]interface{}{
		"Username":          payload.Username,
		"ResetPasswordLink": fmt.Sprintf("%s/api/v1/auth/reset-password/%s", PROJECT_DOMAIN, payload.PasswordResetToken),
	}

	if err := s.mailer.SendMailWithTemplate([]string{payload.Email}, "Reset Your Password", "templates/forgot-password.tmpl", data); err != nil {
		s.log.WithFields(logrus.Fields{"email": payload.Email, "error": err}).Error("Failed to Send Forgot Password Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Forgot Password Email Sent Successfully!")
	s.log.WithFields(logrus.Fields{"email": payload.Email}).Info("Forgot Password Email Sent Successfully!")
}

// handleUserPasswordResetMessage function.
// This function is used to process user password reset messages.
//
// Parameters:
//   - message ([]byte): The message.
func (s *MailerService) handleUserPasswordResetMessage(message []byte) {
	tracer := otel.Tracer("mailer-service")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "mailerGRPCServer.consumeUserPasswordResetMessage")
	defer span.End()

	var payload UserPasswordResetEventPayload
	if err := json.Unmarshal(message, &payload); err != nil {
		s.log.WithFields(logrus.Fields{"message": string(message), "error": err}).Error("Failed to Unmarshal Message!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	s.log.WithFields(logrus.Fields{"email": payload.Email}).Info("Sending Password Reset Email...")

	data := map[string]interface{}{
		"Username":  payload.Username,
		"LoginLink": fmt.Sprintf("%s/api/v1/auth/login", PROJECT_DOMAIN),
	}

	if err := s.mailer.SendMailWithTemplate([]string{payload.Email}, "Password Reset Successfully", "templates/password-reset.tmpl", data); err != nil {
		s.log.WithFields(logrus.Fields{"email": payload.Email, "error": err}).Error("Failed to Send Password Reset Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return
	}

	span.SetStatus(otlpcodes.Ok, "Password Reset Email Sent Successfully!")
	s.log.WithFields(logrus.Fields{"email": payload.Email}).Info("Password Reset Email Sent Successfully!")
}

// Run function.
// This function is used to run the mailer service.
//
// Parameters:
//   - ctx (context.Context): The context.
//
// Returns:
//   - error: The error.
func (s *MailerService) Run(ctx context.Context) error {
	go s.consumeMessages(ctx, USER_REGISTERED_QUEUE, s.handleUserRegisteredMessage)
	go s.consumeMessages(ctx, USER_ACTIVATED_QUEUE, s.handleUserActivatedMessage)
	go s.consumeMessages(ctx, USER_FORGOT_PASSWORD_QUEUE, s.handleUserForgotPasswordMessage)
	go s.consumeMessages(ctx, USER_PASSWORD_RESET_QUEUE, s.handleUserPasswordResetMessage)

	<-ctx.Done()
	s.log.Info("Mailer Service Stopped!")

	return nil
}

// GracefulShutdown function.
// This function is used to shutdown the mailer service gracefully.
//
// Parameters:
//   - log (*logrus.Logger): The logger.
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
func GracefulShutdown(log *logrus.Logger, rabbitMQ *rabbitmqBroker.RabbitMQAdapter) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan
	log.Warn("Shutting Down Gracefully...")

	if err := rabbitMQ.Close(); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close RabbitMQ Connection!")
	}
	log.Info("Disconnected from RabbitMQ!")

	log.Info("Server Stopped!")
}
