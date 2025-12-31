package auth_controller

import (
	authProto "common/proto/auth"
	"errors"
	"gateway/models/auth_models"
	"net/http"

	otlpcodes "go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

// LoginAdmin 		godoc
// @Summary 		Login Admin
// @Description 	Login an Admin to the Platform
// @Tags 			Auth
// @Accept 			json
// @Produce 		json
// @Param 			payload body auth_models.UserLoginPayload true "Admin Login Payload"
// @Success 		200 {object} auth_models.UserLoginSuccessResponse "Successfully logged in admin"
// @Failure 		400 {object} auth_models.UserErrorResponse "Failed to login the admin"
// @Failure 		500 {object} auth_models.UserErrorResponse "Failed to login the admin"
// @Router 			/auth/admin/login [post]
func (ac *AuthController) LoginAdmin(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.LoginAdmin")
	defer span.End()

	var req auth_models.UserLoginPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Login Admin!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserErrorResponse{
			Message: "Failed to Login Admin!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Login Admin!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserErrorResponse{
			Message: "Failed to Login Admin!",
			Error:   err.Error(),
		})
		return
	}

	res, error := ac.client.LoginAdmin(ctx, req.Identifier, req.Password)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Login Admin!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserErrorResponse{
			Message: "Failed to Login Admin!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Admin Logged In Successfully!")
	c.JSON(http.StatusOK, auth_models.UserLoginSuccessResponse{
		Message: "Admin Logged In Successfully!",
		User: auth_models.User{
			ID:           res.User.Id,
			Name:         res.User.Name,
			Username:     res.User.Username,
			Email:        res.User.Email,
			Phone:        res.User.Phone,
			IsSuperAdmin: res.User.IsSuperAdmin,
			IsActive:     res.User.IsActive,
			CreatedAt:    res.User.CreatedAt.AsTime(),
			UpdatedAt:    res.User.UpdatedAt.AsTime(),
		},
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

func (ac *AuthController) LogoutAdmin(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.LogoutAdmin")
	defer span.End()

	var req auth_models.UserLogoutPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Logout Admin!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserErrorResponse{
			Message: "Failed to Logout Admin!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Logout Admin!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserErrorResponse{
			Message: "Failed to Logout Admin!",
			Error:   err.Error(),
		})
		return
	}

	error := ac.client.LogoutAdmin(ctx, req.Identifier)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Logout Admin!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserErrorResponse{
			Message: "Failed to Logout Admin!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Admin Logged Out Successfully!")
	c.JSON(http.StatusOK, auth_models.UserLogoutSuccessResponse{
		Message: "Admin Logged Out Successfully!",
	})
}

func (ac *AuthController) GrantCompanyAccess(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GrantCompanyAccess")
	defer span.End()

	var req auth_models.GrantCompanyAccessPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Grant Company Access!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.GrantCompanyAccessErrorResponse{
			Message: "Failed to Grant Company Access!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Grant Company Access!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.GrantCompanyAccessErrorResponse{
			Message: "Failed to Grant Company Access!",
			Error:   err.Error(),
		})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		ac.logger.Error("Failed to Grant Company Access!")

		span.RecordError(errors.New("user not found"))
		span.SetStatus(otlpcodes.Error, "user not found")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "user not found",
		})
		return
	}

	authUser := user.(*authProto.User)
	if !authUser.IsSuperAdmin {
		ac.logger.Error("Failed to Grant Company Access!")

		span.RecordError(errors.New("user is not an admin"))
		span.SetStatus(otlpcodes.Error, "user is not an admin")

		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
			Message: "Unauthorized!",
			Error:   "only admins can grant access to company",
		})
		return
	}

	res, err := ac.client.GrantCompanyAccess(ctx, req.UserID, req.CompanyIDs)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Grant Company Access!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, auth_models.GrantCompanyAccessErrorResponse{
			Message: "Failed to Grant Company Access!",
			Error:   err.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Company Access Granted Successfully!")
	c.JSON(http.StatusOK, auth_models.GrantCompanyAccessSuccessResponse{
		Message: "Company Access Granted Successfully!",
		User: auth_models.User{
			ID:           res.Id,
			Name:         res.Name,
			Username:     res.Username,
			Email:        res.Email,
			Phone:        res.Phone,
			Incentive:    res.Incentive,
			IsActive:     res.IsActive,
			IsSuperAdmin: res.IsSuperAdmin,
			CompanyIds:   res.CompanyIds,
			CreatedAt:    res.CreatedAt.AsTime(),
			UpdatedAt:    res.UpdatedAt.AsTime(),
		},
	})
}
