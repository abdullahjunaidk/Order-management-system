package main

import (
	rabbitmqBroker "common/broker/rabbitmq"
	mongoDatabase "common/database/mongo"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"common/helpers/logger"
	jaegerMonitoring "common/monitoring/jaeger"
	"context"
	"os"
	"os/signal"
	"syscall"

	inventorySrc "inventory/src"
	productSrc "product/src"

	"order/src"

	"github.com/sirupsen/logrus"
)

var (
	// Service Configuration.
	SERVICE_NAME = env.GetEnv("SERVICE_NAME", "order-service")

	// Database Configuration.
	DATABASE_CONNECTION_URI = env.GetEnv("DATABASE_CONNECTION_URI", "mongodb://localhost:27017")
	DATABASE_NAME           = env.GetEnv("DATABASE_NAME", "order")

	// Server Configuration.
	GRPC_SERVER_PORT = env.GetEnvAsInt("GRPC_SERVER_PORT", 50051)

	// Consul Configuration.
	CONSUL_ADDRESS = env.GetEnv("CONSUL_ADDRESS", "localhost:8500")

	// RabbitMQ Configuration.
	RABBITMQ_ADDRESS = env.GetEnv("RABBITMQ_ADDRESS", "amqp://root:password@localhost:5672/%2f")

	// Jaeger Configuration
	OTLP_ENDPOINT = env.GetEnv("OTLP_ENDPOINT", "localhost:4317")
	OTLP_PROTOCOL = env.GetEnv("OTLP_PROTOCOL", "grpc")
	OTLP_INSECURE = env.GetEnv("OTLP_INSECURE", "true")

	// Instance Configuration.
	INSTANCE_ID           = env.GetEnv("INSTANCE_ID", "8727ed1f-bd9e-4fe9-9978-3e20acac4601")
	INSTANCE_HOST         = env.GetEnv("INSTANCE_HOST", "localhost")
	INSTANCE_PORT         = env.GetEnvAsInt("INSTANCE_PORT", 50051)
	HEALTH_CHECK_URL      = env.GetEnv("HEALTH_CHECK_URL", "")
	HEALTH_CHECK_INTERVAL = env.GetEnv("HEALTH_CHECK_INTERVAL", "10s")

	// Microservice Configuration.
	PRODUCT_SERVICE_NAME   = env.GetEnv("PRODUCT_SERVICE_NAME", "product-service")
	INVENTORY_SERVICE_NAME = env.GetEnv("INVENTORY_SERVICE_NAME", "inventory-service")

	// RabbitMQ Queues
	ORDER_PENDING_CANCELLED_QUEUE = env.GetEnv("ORDER_PENDING_CANCELLED_QUEUE", "order.pendingOrder.cancelled")
	ORDER_PAID_CANCELLED_QUEUE    = env.GetEnv("ORDER_PAID_CANCELLED_QUEUE", "order.paidOrder.cancelled")
)

func main() {
	log := logger.NewLogger()

	//consul registry - connect to consul
	consulRegistry, err := consulDiscovery.NewConsulRegistry(CONSUL_ADDRESS)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to Consul!")
	}
	log.Info("Connected to Consul!")

	//mongo adapter - connect to mongodb
	mongoAdapter, err := mongoDatabase.NewMongoDBAdapter(DATABASE_CONNECTION_URI, DATABASE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to Database!")
	}
	log.Info("Connected to Database!")

	//rabbitmq adapter
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

	orderStore := src.NewOrderStore(*mongoAdapter)

	//jaeger tracer - initialize global tracer
	jaegerTracer, err := jaegerMonitoring.NewJaegerTracer(SERVICE_NAME, OTLP_ENDPOINT, OTLP_PROTOCOL, OTLP_INSECURE)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Initialize Global Tracer!")
	}
	defer func() {
		jaegerTracer.Close()
		log.Info("Global Tracer Shutdown!")
	}()
	log.Info("Initialized Global Tracer!")

	//product service
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


	//inventory service
	inventoryServiceAddress, err := consulRegistry.GetServiceAddress(INVENTORY_SERVICE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Inventory Service Address!")
	}

	inventoryClient, err := inventorySrc.NewInventoryClient(inventoryServiceAddress, SERVICE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Inventory Service Client!")
	}
	defer func() {
		inventoryClient.Close()
		log.Info("Disconnected from Inventory Service!")
	}()

	//order service - create order service
	orderService := src.NewOrderService(orderStore, productClient, inventoryClient, rabbitMQ)

	//server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := src.ListenAndServeGRPC(orderService, GRPC_SERVER_PORT); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Start gRPC Server!")
		}
	}()

	//register service with consul
	go func() {
		log.Info("Registering Service with Consul...")
		if err := consulRegistry.RegisterService(SERVICE_NAME, INSTANCE_ID, INSTANCE_HOST, INSTANCE_PORT, true, HEALTH_CHECK_URL, HEALTH_CHECK_INTERVAL); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Register Service with Consul!")
		}
		log.Info("Service Registered with Consul!")
	}()

	<-ctx.Done()

	stop()
	log.Warn("Shutting Down Gracefully...")

	//deregister service from consul
	log.Warn("Deregistering Service from Consul...")
	if err := consulRegistry.DeregisterService(INSTANCE_ID); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Deregister Service from Consul!")
	}
	log.Info("Service Deregistered from Consul!")

	//disconnect from mongodb
	if err := mongoAdapter.Disconnect(context.Background()); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Disconnect from MongoDB!")
	}
	log.Info("Disconnected from MongoDB!")

	log.Info("Server Stopped!")
}
