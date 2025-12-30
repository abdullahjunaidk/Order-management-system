package auth_controller

import (
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

// func (ac *AuthController) EmployeeGrantAccess(c *gin.Context) {
// 	tracer := otel.Tracer("gateway-service")

// 	ctx := c.Request.Context()
// 	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.EmployeeGrantAccess")
// 	defer span.End()

// 	var req auth_models.EmployeeBrandAccessRegisterPayload
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Grant Employee Access!")

// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())

// 		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{
// 			Message: "Failed to Grant Employee Access!",
// 			Error:   err.Error(),
// 		})
// 		return
// 	}

// 	if err := Validate.Struct(req); err != nil {
// 		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Grant Employee Access!")

// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())

// 		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{
// 			Message: "Failed to Grant Employee Access!",
// 			Error:   err.Error(),
// 		})
// 		return
// 	}

// 	admin, exists := c.Get("user")
// 	if !exists {
// 		ac.logger.Error("Failed to Grant Employee Access!")

// 		span.RecordError(errors.New("user not found"))
// 		span.SetStatus(otlpcodes.Error, "user not found")

// 		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
// 			Message: "Unauthorized!",
// 			Error:   "user not found",
// 		})
// 		return
// 	}

// 	authUser := admin.(*authProto.Employee)
// 	if authUser.Role != "admin" {
// 		ac.logger.Error("Failed to Grant Employee Access!")

// 		span.RecordError(errors.New("user is not an admin"))
// 		span.SetStatus(otlpcodes.Error, "user is not an admin")

// 		c.JSON(http.StatusUnauthorized, auth_models.UnauthorizedErrorResponse{
// 			Message: "Unauthorized!",
// 			Error:   "only admins can Grant access to employee",
// 		})
// 		return
// 	}

// 	var access []*authProto.BrandAccess
// 	for _, reqAccess := range req.Access {
// 		access = append(access, &authProto.BrandAccess{
// 			BrandId: reqAccess.BrandID,
// 			RoleId:  reqAccess.Roles,
// 		})
// 	}

// 	res, error := ac.client.EmployeeGrantAccess(ctx, req.EmployeeID, access)
// 	if error != nil {
// 		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Grant Employee Access!")

// 		span.RecordError(error)
// 		span.SetStatus(otlpcodes.Error, error.Error())

// 		c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{
// 			Message: "Failed to Grant Employee Access!",
// 			Error:   error.Error(),
// 		})
// 		return
// 	}

// 	var ResAccess []auth_models.BrandAccessDTO
// 	for _, access := range res.Access {
// 		ResAccess = append(ResAccess, auth_models.BrandAccessDTO{
// 			BrandID: access.BrandId,
// 			Roles:   access.RoleId,
// 		})
// 	}

// 	span.SetStatus(otlpcodes.Ok, "Employee Access Granted Successfully!")
// 	c.JSON(http.StatusOK, auth_models.EmployeeGrandAccessSuccessResponse{
// 		Message: "Employee Access Granted Successfully!",
// 		Employee: auth_models.Employee{
// 			ID:        res.Id,
// 			UserName:  res.Username,
// 			Name:      res.Name,
// 			Phone:     res.Phone,
// 			Email:     res.Email,
// 			Incentive: res.Incentive,
// 			Permission: res.Permissions,
// 			IsActive:  res.IsActive,
// 			Role:      res.Role,
// 			CreatedAt: res.CreatedAt.AsTime(),
// 			UpdatedAt: res.UpdatedAt.AsTime(),
// 		},
// 	})
// }

// func (ac *AuthController) GrandAccess(c *gin.Context) {
// 	tracer := otel.Tracer("gateway-service")

// 	ctx := c.Request.Context()
// 	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GrandAccess")
// 	defer span.End()

// 	var req auth_models.EmployeeBrandAccessRegisterPayload
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Grand Access!")

// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())

// 		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{
// 			Message: "Failed to Grand Access!",
// 			Error:   err.Error(),
// 		})
// 		return
// 	}

// 	if err := Validate.Struct(req); err != nil {
// 		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Grand Access!")

// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())

// 		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{
// 			Message: "Failed to Grand Access!",
// 			Error:   err.Error(),
// 		})
// 		return
// 	}

// 	res, error := ac.client.GrandAccess(ctx, req.EmployeeID, req.Access)
// 	if error != nil {
// 		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Grand Access!")

// 		span.RecordError(error)
// 		span.SetStatus(otlpcodes.Error, error.Error())

// 		c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{
// 			Message: "Failed to Grand Access!",
// 			Error:   error.Error(),
// 		})
// 		return
// 	}

// 	span.SetStatus(otlpcodes.Ok, "Access Granted Successfully!")
// 	c.JSON(http.StatusOK, auth_models.ErrorResponse{
// 		Message: "Access Granted Successfully!",
// 	})
// }
