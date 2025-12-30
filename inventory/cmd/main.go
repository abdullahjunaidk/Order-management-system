package main

import (
	mongoDatabase "common/database/mongo"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"common/helpers/logger"
	jaegerMonitoring "common/monitoring/jaeger"
	"context"
	"os"
	"os/signal"
	"syscall"

	"inventory/src"

	"github.com/sirupsen/logrus"
)

var (
	// Service Configuration.
	SERVICE_NAME = env.GetEnv("SERVICE_NAME", "inventory-service")

	// Database Configuration.
	DATABASE_CONNECTION_URI = env.GetEnv("DATABASE_CONNECTION_URI", "mongodb://localhost:27017")
	DATABASE_NAME           = env.GetEnv("DATABASE_NAME", "inventory")

	// Server Configuration.
	GRPC_SERVER_PORT = env.GetEnvAsInt("GRPC_SERVER_PORT", 50051)

	// Consul Configuration.
	CONSUL_ADDRESS = env.GetEnv("CONSUL_ADDRESS", "localhost:8500")

	// Jaeger Configuration
	OTLP_ENDPOINT = env.GetEnv("OTLP_ENDPOINT", "localhost:4317")
	OTLP_PROTOCOL = env.GetEnv("OTLP_PROTOCOL", "grpc")
	OTLP_INSECURE = env.GetEnv("OTLP_INSECURE", "true")

	// Instance Configuration.
	INSTANCE_ID           = env.GetEnv("INSTANCE_ID", "ee6085e0-6bf2-4393-bb40-6d4d92cc1c97")
	INSTANCE_HOST         = env.GetEnv("INSTANCE_HOST", "localhost")
	INSTANCE_PORT         = env.GetEnvAsInt("INSTANCE_PORT", 50051)
	HEALTH_CHECK_URL      = env.GetEnv("HEALTH_CHECK_URL", "")
	HEALTH_CHECK_INTERVAL = env.GetEnv("HEALTH_CHECK_INTERVAL", "10s")
)

func main() {
	log := logger.NewLogger()

	consulRegistry, err := consulDiscovery.NewConsulRegistry(CONSUL_ADDRESS)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to Consul!")
	}
	log.Info("Connected to Consul!")

	mongoAdapter, err := mongoDatabase.NewMongoDBAdapter(DATABASE_CONNECTION_URI, DATABASE_NAME)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to Database!")
	}
	log.Info("Connected to Database!")

	inventoryStore := src.NewInventoryStore(*mongoAdapter)

	jaegerTracer, err := jaegerMonitoring.NewJaegerTracer(SERVICE_NAME, OTLP_ENDPOINT, OTLP_PROTOCOL, OTLP_INSECURE)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Initialize Global Tracer!")
	}
	defer func() {
		jaegerTracer.Close()
		log.Info("Global Tracer Shutdown!")
	}()
	log.Info("Initialized Global Tracer!")

	inventoryService := src.NewInventoryService(inventoryStore)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := src.ListenAndServeGRPC(inventoryService, GRPC_SERVER_PORT); err != nil {
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

	<-ctx.Done()

	stop()
	log.Warn("Shutting Down Gracefully...")

	log.Warn("Deregistering Service from Consul...")
	if err := consulRegistry.DeregisterService(INSTANCE_ID); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Deregister Service from Consul!")
	}
	log.Info("Service Deregistered from Consul!")

	if err := mongoAdapter.Disconnect(context.Background()); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Disconnect from MongoDB!")
	}
	log.Info("Disconnected from MongoDB!")

	log.Info("Server Stopped!")
}
