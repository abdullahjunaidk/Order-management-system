package auth_controller

import (
	"common/helpers/password"
	"errors"
	"gateway/models/auth_models"
	"net/http"

	otlpcodes "go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

// RegisterEmployee godoc
// @Summary Register a new Employee
// @Description Register a new Employee to the Platform
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body auth_models.EmployeeRegisterPayload true "Employee Register Payload"
// @Success 200 {object} auth_models.EmployeeRegisterSuccessResponse "Successfully registered employee"
// @Failure 400 {object} auth_models.EmployeeRegisterErrorResponse "Failed to register the employee"
// @Failure 500 {object} auth_models.EmployeeRegisterErrorResponse "Failed to register the employee"
// @Router /auth/employee/register [post]
func (ac *AuthController) RegisterUser(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.RegisterEmployee")
	defer span.End()

	var req auth_models.UserRegisterPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Register Employee!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserRegisterErrorResponse{
			Message: "Failed to Register Employee!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Register Employee!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserRegisterErrorResponse{
			Message: "Failed to Register Employee!",
			Error:   err.Error(),
		})
		return
	}

	hashedPassword, err := password.HashPassword(req.PasswordHash, 10)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Register Employee!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserRegisterErrorResponse{
			Message: "Failed to Register Employee!",
			Error:   err.Error(),
		})
		return
	}

	res, error := ac.client.RegisterUser(ctx, req.Name, req.UserName, req.Email, hashedPassword, req.Phone, req.CompanyIds)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Register Employee!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserRegisterErrorResponse{
			Message: "Failed to Register Employee!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Employee Registered Successfully!")
	c.JSON(http.StatusOK, auth_models.UserRegisterSuccessResponse{
		Message: "Employee Registered Successfully!",
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

// ActivateEmployee 		godoc
// @Summary 			Activate an Employee
// @Description 		Activate an Employee in the Platform
// @Tags 				Auth
// @Accept 				json
// @Produce 			json
// @Param 				activationToken path string true "Activation Token"
// @Success 			200 {object} auth_models.UserActivateSuccessResponse "Successfully activated employee"
// @Failure 			400 {object} auth_models.UserActivateErrorResponse "Failed to activate the employee"
// @Failure 			500 {object} auth_models.UserActivateErrorResponse "Failed to activate the employee"
// @Router 				/auth/employee/activate/{activationToken} [get]
func (ac *AuthController) ActivateUser(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.ActivateUser")
	defer span.End()

	activationToken := c.Param("activationToken")

	res, err := ac.client.ActivateUser(ctx, activationToken)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Activate User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserActivateErrorResponse{
			Message: "Failed to Activate User!",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, auth_models.UserActivateSuccessResponse{
		Message: "User Activated Successfully!",
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

func (ac *AuthController) LoginUser(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.LoginUser")
	defer span.End()

	var req auth_models.UserLoginPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Login User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserErrorResponse{
			Message: "Failed to Login User!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Login User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserErrorResponse{
			Message: "Failed to Login User!",
			Error:   err.Error(),
		})
		return
	}

	res, error := ac.client.LoginUser(ctx, req.Identifier, req.Password)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Login User!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserErrorResponse{
			Message: "Failed to Login User!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "User Logged In Successfully!")
	c.JSON(http.StatusOK, auth_models.UserLoginSuccessResponse{
		Message: "User Logged In Successfully!",
		User: auth_models.User{
			ID:           res.User.Id,
			Name:         res.User.Name,
			Username:     res.User.Username,
			Email:        res.User.Email,
			Phone:        res.User.Phone,
			Incentive:    res.User.Incentive,
			IsActive:     res.User.IsActive,
			IsSuperAdmin: res.User.IsSuperAdmin,
			CompanyIds:   res.User.CompanyIds,
			CreatedAt:    res.User.CreatedAt.AsTime(),
			UpdatedAt:    res.User.UpdatedAt.AsTime(),
		},
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

func (ac *AuthController) LogoutUser(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.LogoutUser")
	defer span.End()

	var req auth_models.UserLogoutPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Logout User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserErrorResponse{
			Message: "Failed to Logout User!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Logout User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserErrorResponse{
			Message: "Failed to Logout User!",
			Error:   err.Error(),
		})
		return
	}

	error := ac.client.LogoutUser(ctx, req.Identifier)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Logout User!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserErrorResponse{
			Message: "Failed to Logout User!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "User Logged Out Successfully!")
	c.JSON(http.StatusOK, auth_models.UserLogoutSuccessResponse{
		Message: "User Logged Out Successfully!",
	})
}

// RefreshAccessToken 		godoc
// @Summary 				Refresh Access Token
// @Description 			Refresh Access Token for a User
// @Tags 					Auth
// @Accept 					json
// @Produce 				json
// @Param 					payload body models.UserRefreshAccessTokenPayload true "User Refresh Access Token Payload"
// @Success 				200 {object} models.UserRefreshAccessTokenSuccessResponse "Successfully refreshed user access token"
// @Failure 				400 {object} models.UserRefreshAccessTokenErrorResponse "Failed to refresh user access token"
// @Failure 				500 {object} models.UserRefreshAccessTokenErrorResponse "Failed to refresh user access token"
// @Router 					/auth/refresh-access-token [post]
func (ac *AuthController) RefreshAccessToken(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.RefreshAccessToken")
	defer span.End()

	var req auth_models.UserRefreshAccessTokenPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Refresh User Access Token!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserRefreshAccessTokenErrorResponse{
			Message: "Failed to Refresh User Access Token!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Refresh User Access Token!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.UserRefreshAccessTokenErrorResponse{
			Message: "Failed to Refresh User Access Token!",
			Error:   err.Error(),
		})
		return
	}

	res, error := ac.client.RefreshAccessToken(ctx, req.RefreshToken)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Refresh User Access Token!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserRefreshAccessTokenErrorResponse{
			Message: "Failed to Refresh User Access Token!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "User Access Token Refreshed Successfully!")
	c.JSON(http.StatusOK, auth_models.UserRefreshAccessTokenSuccessResponse{
		Message: "User Access Token Refreshed Successfully!",
		User: auth_models.User{
			ID:           res.User.Id,
			Name:         res.User.Name,
			Username:     res.User.Username,
			Email:        res.User.Email,
			Phone:        res.User.Phone,
			Incentive:    res.User.Incentive,
			IsActive:     res.User.IsActive,
			IsSuperAdmin: res.User.IsSuperAdmin,
			CompanyIds:   res.User.CompanyIds,
			CreatedAt:    res.User.CreatedAt.AsTime(),
			UpdatedAt:    res.User.UpdatedAt.AsTime(),
		},
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	})
}

// GetUserByID godoc
// @Summary 			Get User by ID
// @Description 		Get a User by their ID
// @Tags 				Auth
// @Accept 				json
// @Produce 			json
// @Param 				userId path string true "User ID"
// @Success 			200 {object} auth_models.User "Successfully retrieved user"
// @Failure 			400 {object} auth_models.ErrorResponse "Failed to get user"
// @Failure 			500 {object} auth_models.ErrorResponse "Failed to get user"
// @Router 				/auth/user/{userId} [get]
func (ac *AuthController) GetUserByID(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")
	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GetUserByID")
	defer span.End()

	userId := c.Param("userId")
	if userId == "" {
		span.RecordError(errors.New("user ID is required"))
		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{Message: "User ID is required"})
		return
	}

	res, err := ac.client.GetUserByID(ctx, userId)
	if err != nil {
		span.RecordError(err)
		c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{Message: "Failed to get user", Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, auth_models.User{
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
	})
}
