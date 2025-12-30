package auth_controller

import (
	authProto "common/proto/auth"

	"errors"
	"gateway/models/auth_models"
	"net/http"
	"strings"

	otlpcodes "go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

func (ac *AuthController) RegisterRole(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.RegisterRole")
	defer span.End()

	var req auth_models.RegisterRolePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Register Role!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{
			Message: "Failed to Register Role!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Register Role!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{
			Message: "Failed to Register Role!",
			Error:   err.Error(),
		})
		return
	}

	role := strings.ToLower(req.Name)

	res, error := ac.client.RegisterRole(ctx, role)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Register Role!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{
			Message: "Failed to Register Role!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Role Registered Successfully!")
	c.JSON(http.StatusOK, auth_models.RegisterRoleResponse{
		Message: "Role Registered Successfully!",
		Role: auth_models.Role{
			ID:          res.Id,
			Name:        res.Name,
			Description: res.Description,
			CreatedAt:   res.CreatedAt.AsTime(),
			UpdatedAt:   res.UpdatedAt.AsTime(),
		},
	})
}

func (ac *AuthController) GetRoleByID(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GetRoleByID")
	defer span.End()

	roleId := c.Param("roleId")
	if roleId == "" {
		ac.logger.WithFields(logrus.Fields{"error": "Role ID is required!"}).Error("Failed to Get Role by ID!")

		span.RecordError(errors.New("Role ID is required!"))
		span.SetStatus(otlpcodes.Error, "Role ID is required!")

		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{
			Message: "Failed to Get Role by ID!",
			Error:   "Role ID is required!",
		})
		return
	}

	res, error := ac.client.GetRoleByID(ctx, roleId)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Get Role by ID!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{
			Message: "Failed to Get Role by ID!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Role Retrieved Successfully!")
	c.JSON(http.StatusOK, auth_models.Role{
		ID:          res.Id,
		Name:        res.Name,
		Description: res.Description,
		CreatedAt:   res.CreatedAt.AsTime(),
		UpdatedAt:   res.UpdatedAt.AsTime(),
	})
}

func (ac *AuthController) GetRoles(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GetRoles")
	defer span.End()

	admin, exists := c.Get("employee")
	if !exists {
		ac.logger.WithFields(logrus.Fields{"error": "admin not found"}).Error("Failed to Get Roles!")

		span.RecordError(errors.New("admin not found"))
		span.SetStatus(otlpcodes.Error, "admin not found")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "admin not found",
		})
		return
	}

	authAdmin := admin.(*authProto.User)
	if !authAdmin.IsSuperAdmin {
		ac.logger.WithFields(logrus.Fields{"error": "user is not a admin"}).Error("Failed to Get Roles!")

		span.RecordError(errors.New("user is not a admin"))
		span.SetStatus(otlpcodes.Error, "user is not a admin")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "user is not a admin",
		})
		return
	}

	res, error := ac.client.GetRoles(ctx)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Get Roles!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{
			Message: "Failed to Get Roles!",
			Error:   error.Error(),
		})
		return
	}

	roles := []auth_models.Role{}
	for _, role := range res.Roles {
		roles = append(roles, auth_models.Role{
			ID:          role.Id,
			Name:        role.Name,
			Description: role.Description,
			CreatedAt:   role.CreatedAt.AsTime(),
			UpdatedAt:   role.UpdatedAt.AsTime(),
		})
	}

	span.SetStatus(otlpcodes.Ok, "Roles Retrieved Successfully!")
	c.JSON(http.StatusOK, roles)
}
