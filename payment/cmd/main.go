package main

import (
	rabbitmqBroker "common/broker/rabbitmq"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"common/helpers/logger"
	jaegerMonitoring "common/monitoring/jaeger"
	"context"
	"os"
	"os/signal"
	"syscall"

	authSrc "auth/src"
	orderSrc "order/src"
	"payment/src"
	productSrc "product/src"

	"github.com/sirupsen/logrus"
)

var (
	// Service Configuration.
	SERVICE_NAME = env.GetEnv("SERVICE_NAME", "payment-service")

	// Server Configuration.
	GRPC_SERVER_PORT = env.GetEnvAsInt("GRPC_SERVER_PORT", 50052)

	// Consul Configuration.
	CONSUL_ADDRESS = env.GetEnv("CONSUL_ADDRESS", "localhost:8500")

	// RabbitMQ Configuration.
	RABBITMQ_ADDRESS = env.GetEnv("RABBITMQ_ADDRESS", "amqp://root:password@localhost:5672/%2f")

	// Jaeger Configuration
	OTLP_ENDPOINT = env.GetEnv("OTLP_ENDPOINT", "localhost:4317")
	OTLP_PROTOCOL = env.GetEnv("OTLP_PROTOCOL", "grpc")
	OTLP_INSECURE = env.GetEnv("OTLP_INSECURE", "true")

	// Instance Configuration.
	INSTANCE_ID           = env.GetEnv("INSTANCE_ID", "b0188a67-ffc8-402f-a41b-1ada65435346")
	INSTANCE_HOST         = env.GetEnv("INSTANCE_HOST", "localhost")
	INSTANCE_PORT         = env.GetEnvAsInt("INSTANCE_PORT", 50052)
	HEALTH_CHECK_URL      = env.GetEnv("HEALTH_CHECK_URL", "")
	HEALTH_CHECK_INTERVAL = env.GetEnv("HEALTH_CHECK_INTERVAL", "10s")

	// Microservice Configuration
	AUTH_SERVICE_NAME    = env.GetEnv("AUTH_SERVICE_NAME", "auth-service")
	PRODUCT_SERVICE_NAME = env.GetEnv("PRODUCT_SERVICE_NAME", "product-service")
	ORDER_SERVICE_NAME   = env.GetEnv("ORDER_SERVICE_NAME", "order-service")
)

func main() {
	log := logger.NewLogger()

	consulRegistry, err := consulDiscovery.NewConsulRegistry(CONSUL_ADDRESS)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to Consul!")
	}
	log.Info("Connected to Consul!")

	rabbitMQ, err := rabbitmqBroker.NewRabbitMQAdapter(RABBITMQ_ADDRESS)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to RabbitMQ!")
	}
	defer func() {
		if err := rabbitMQ.Close(); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close RabbitMQ Connection!")
		}
		log.Info("Disconnected from RabbitMQ!")
	}()
	log.Info("Connected to RabbitMQ!")

	jaegerTracer, err := jaegerMonitoring.NewJaegerTracer(SERVICE_NAME, OTLP_ENDPOINT, OTLP_PROTOCOL, OTLP_INSECURE)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Initialize Global Tracer!")
	}
	defer func() {
		jaegerTracer.Close()
		log.Info("Global Tracer Shutdown!")
	}()
	log.Info("Initialized Global Tracer!")

	authServiceAddress, err := consulRegistry.GetServiceAddress(AUTH_SERVICE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Auth Service Address!")
	}

	authClient, err := authSrc.NewAuthClient(authServiceAddress, SERVICE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Auth Service Client!")
	}
	defer func() {
		authClient.Close()
		log.Info("Disconnected from Auth Service!")
	}()

	productServiceAddress, err := consulRegistry.GetServiceAddress(PRODUCT_SERVICE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Product Service Address!")
	}

	productClient, err := productSrc.NewProductClient(productServiceAddress, SERVICE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Product Service Client!")
	}
	defer func() {
		productClient.Close()
		log.Info("Disconnected from Product Service!")
	}()

	orderServiceAddress, err := consulRegistry.GetServiceAddress(ORDER_SERVICE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Order Service Address!")
	}

	orderClient, err := orderSrc.NewOrderClient(orderServiceAddress, SERVICE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Order Service Client!")
	}
	defer func() {
		orderClient.Close()
		log.Info("Disconnected from Order Service!")
	}()

	stripeAdapter, err := src.NewStripeAdapter(log)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Stripe Adapter!")
	}

	paymentService := src.NewPaymentService(rabbitMQ, log, authClient, productClient, orderClient, stripeAdapter)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := src.ListenAndServeGRPC(GRPC_SERVER_PORT); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Start gRPC Server!")
		}
	}()

	go func() {
		log.Info("Registering Service with Consul...")
		if err := consulRegistry.RegisterService(SERVICE_NAME, INSTANCE_ID, INSTANCE_HOST, INSTANCE_PORT, true, HEALTH_CHECK_URL, HEALTH_CHECK_INTERVAL); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Register Service with Consul!")
		}
		log.Info("Service Registered with Consul!")
	}()

	defer src.GracefulShutdown(log, rabbitMQ, productClient)

	if err := paymentService.Run(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()

	stop()
	log.Warn("Shutting Down Gracefully...")

	log.Warn("Deregistering Service from Consul...")
	if err := consulRegistry.DeregisterService(INSTANCE_ID); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Deregister Service from Consul!")
	}
	log.Info("Service Deregistered from Consul!")

	log.Info("Server Stopped!")
}
