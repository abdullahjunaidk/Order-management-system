package middlewares

import (
	authSrc "auth/src"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"context"
	"gateway/models/auth_models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	// Microservices Configuration
	AUTH_SERVICE_NAME = env.GetEnv("AUTH_SERVICE_NAME", "auth-service")
)

// AuthMiddleware is a middleware function to authenticate user requests using JWT tokens from cookies.
// It checks for access token and refresh token cookies, verifies them, and sets the user in the context.
// It also handles access token refreshing using refresh token if access token is expired.
//
// Parameters:
//   - registry (*consulDiscovery.ConsulRegistry): Consul registry instance for service discovery.
//   - logger (*logrus.Logger): Logrus logger instance for logging.
//
// Returns:
//   - gin.HandlerFunc: Gin middleware handler function.
func AuthMiddleware(registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.WithFields(logrus.Fields{"error": "authorization header is required"}).Error("Unauthorized!")
			c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
				Message: "Unauthorized!",
				Error:   "authorization header is required",
			})
			c.Abort()
			return
		}

		authServiceAddress, err := registry.GetServiceAddress(AUTH_SERVICE_NAME)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Auth Service Address!")
			c.Abort()
		}

		authServiceClient, err := authSrc.NewAuthClient(authServiceAddress, "gateway-service")
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Auth Service Client!")
			c.Abort()
		}

		user, err := authServiceClient.VerifyAccessToken(context.Background(), authHeader)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Get User!")
			c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
				Message: "Unauthorized!",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
