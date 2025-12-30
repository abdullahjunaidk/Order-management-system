package main

import (
	redisBroker "common/broker/redis"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"common/helpers/logger"
	"common/monitoring/jaeger"
	jaegerMonitoring "common/monitoring/jaeger"
	"context"
	"fmt"
	_ "gateway/docs"
	"gateway/middlewares"
	"gateway/routes"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	// Service Configuration.
	SERVICE_NAME = env.GetEnv("SERVICE_NAME", "gateway-service")

	// Server Configuration.
	HTTP_SERVER_MODE = env.GetEnv("HTTP_SERVER_MODE", "release")
	HTTP_SERVER_PORT = env.GetEnvAsInt("HTTP_SERVER_PORT", 8080)

	// Consul Configuration.
	CONSUL_ADDRESS = env.GetEnv("CONSUL_ADDRESS", "localhost:8500")

	// Jaeger Configuration
	OTLP_ENDPOINT = env.GetEnv("OTLP_ENDPOINT", "localhost:4318")
	OTLP_PROTOCOL = env.GetEnv("OTLP_PROTOCOL", "http")
	OTLP_INSECURE = env.GetEnv("OTLP_INSECURE", "true")

	// Instance Configuration.
	INSTANCE_ID           = env.GetEnv("INSTANCE_ID", "3168bd6b-fe9a-4908-bff5-e7c368b83fff")
	INSTANCE_HOST         = env.GetEnv("INSTANCE_HOST", "localhost")
	INSTANCE_PORT         = env.GetEnvAsInt("INSTANCE_PORT", 8080)
	HEALTH_CHECK_URL      = env.GetEnv("HEALTH_CHECK_URL", "/api/v1/ping")
	HEALTH_CHECK_INTERVAL = env.GetEnv("HEALTH_CHECK_INTERVAL", "10s")

	// Redis Configuration.
	REDIS_CONNECTION_URI = env.GetEnv("REDIS_CONNECTION_URI", "redis://localhost:6379")
)

// @title           Gopher Social API
// @version         1.0
// @description     This is the API for Gopher Social, a social media platform for Gophers.

// @contact.name   Rohit Vilas Ingole
// @contact.email  rohit.vilas.ingole@gmail.com

// @license.name  MIT License
// @license.url   https://github.com/DataRohit/Order-Management-System/blob/master/license

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	gin.SetMode(HTTP_SERVER_MODE)

	log := logger.NewLogger()

	// Initialize Consul Registry.
	consulRegistry, err := consulDiscovery.NewConsulRegistry(CONSUL_ADDRESS)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to Consul!")
	}
	log.Info("Connected to Consul!")

	// Jaeger Tracer.
	jaegerTracer, err := jaegerMonitoring.NewJaegerTracer(SERVICE_NAME, OTLP_ENDPOINT, OTLP_PROTOCOL, OTLP_INSECURE)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Initialize Global Tracer!")
	}
	defer func() {
		jaegerTracer.Close()
		log.Info("Global Tracer Shutdown!")
	}()
	log.Info("Initialized Global Tracer!")

	// Initialize Redis Adapter.
	redis, err := redisBroker.NewRedisAdapter(REDIS_CONNECTION_URI, log)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to Redis!")
	}
	log.Info("Connected to Redis!")

	// Initialize Jaeger Tracer.
	closer, err := jaeger.InitGlobalTracer(SERVICE_NAME, OTLP_ENDPOINT, OTLP_PROTOCOL)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Initialize Jaeger Tracer!")
	}

	defer func() {
		if err := closer.Close(); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close Jaeger Tracer!")
		}

		if err := redis.Close(); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close Redis Connection!")
		}
		log.Info("Disconnected from Redis!")
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := ListenAndServer(HTTP_SERVER_PORT, consulRegistry, log); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Start HTTP Server!")
		}
	}()

	go func() {
		log.Info("Registering Service with Consul...")
		if err := consulRegistry.RegisterService(SERVICE_NAME, INSTANCE_ID, INSTANCE_HOST, INSTANCE_PORT, false, HEALTH_CHECK_URL, HEALTH_CHECK_INTERVAL); err != nil {
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

	log.Info("Server Stopped!")
}

// ListenAndServer function.
// This function is used to listen and serve the gateway service HTTP server.
//
// Parameters:
//   - port (int): The port.
//   - registry (*consulDiscovery.ConsulRegistry): The consul registry.
//   - log (*logrus.Logger): The logger.
func ListenAndServer(port int, registry *consulDiscovery.ConsulRegistry, log *logrus.Logger) error {
	redis, err := redisBroker.NewRedisAdapter(REDIS_CONNECTION_URI, log)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Connect to Redis!")
	}
	log.Info("Connected to Redis!")

	defer func() {
		if err := redis.Close(); err != nil {
			log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close Redis Connection!")
		}
		log.Info("Disconnected from Redis!")
	}()

	router := gin.New()

	router.Use(middlewares.RealIPMiddleware())
	router.Use(middlewares.LoggerMiddleware(log))
	router.Use(middlewares.RecovererMiddleware(log))
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.TimeoutMiddleware(5 * time.Second))
	router.Use(middlewares.RateLimiterMiddleware(redis.Client, 120, time.Minute, log))

	apiv1 := router.Group("/api/v1")

	apiv1.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	registerRoutes(apiv1, registry, log)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	log.WithFields(logrus.Fields{"port": port}).Info("Starting HTTP Server...")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Serve HTTP Server!")
		return fmt.Errorf("failed to serve: %w", err)
	}
	log.Info("HTTP Server Stopped!")

	return nil
}

// registerRoutes function.
// This function is used to register the routes for the gateway service.
//
// Parameters:
//   - router (*gin.RouterGroup): The router group.
//   - registry (*consulDiscovery.ConsulRegistry): The consul registry.
//   - logger (*logrus.Logger): The logger.
func registerRoutes(router *gin.RouterGroup, registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) {
	routes.AuthRoutes(router, registry, logger)
	routes.ProductRoutes(router, registry, logger)
	routes.InventoryRoutes(router, registry, logger)
	routes.OrderRoutes(router, registry, logger)
}
