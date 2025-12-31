package middlewares

import (
	authSrc "auth/src"
	consulDiscovery "common/discovery/consul"
	authProto "common/proto/auth"
	"context"
	"gateway/models/auth_models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// PermissionMiddleware checks if the user has the required permission for a resource and action
func PermissionMiddleware(registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by AuthMiddleware)
		userCtx, exists := c.Get("user")
		if !exists {
			logger.Error("User not found in context")
			c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
				Message: "Unauthorized",
				Error:   "User not authenticated",
			})
			c.Abort()
			return
		}

		user := userCtx.(*authProto.User)

		// Create Auth Client
		authServiceAddress, err := registry.GetServiceAddress(AUTH_SERVICE_NAME)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Auth Service Address!")
			c.Abort()
			return
		}

		authServiceClient, err := authSrc.NewAuthClient(authServiceAddress, "gateway-service")
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Auth Service Client!")
			c.Abort()
			return
		}
		defer authServiceClient.Close()

		// Check Permission
		allowed, err := authServiceClient.CheckUserPermission(context.Background(), user.Id, resource, action)
		if err != nil {
			logger.WithFields(logrus.Fields{"error": err, "userId": user.Id}).Error("Failed to check permission")
			c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{
				Message: "Internal Server Error",
				Error:   err.Error(),
			})
			c.Abort()
			return
		}

		if !allowed {
			logger.WithFields(logrus.Fields{"userId": user.Id, "resource": resource, "action": action}).Warn("Permission denied")
			c.JSON(http.StatusForbidden, auth_models.ErrorResponse{
				Message: "Permission Denied",
				Error:   "You does not have permission to perform this action",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
