package auth_controller

import (
	authSrc "auth/src"
	consulDiscovery "common/discovery/consul"
	"common/helpers/env"
	"common/helpers/password"
	"errors"
	"gateway/models/auth_models"
	"net/http"

	otlpcodes "go.opentelemetry.io/otel/codes"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

var (
	// Microservices Configuration
	AUTH_SERVICE_NAME = env.GetEnv("AUTH_SERVICE_NAME", "auth-service")

	// Bcrypt Configuration.
	BCRYPT_HASHING_COST = env.GetEnvAsInt("BCRYPT_HASHING_COST", 10)
)

// Initialize a validator instance.
var Validate = validator.New()

// AuthController struct.
// This struct is used to represent an Auth controller.
//
// Attributes:
//   - client (*authSrc.AuthClient): The Auth client.
//   - logger (*logrus.Logger): The logger.
type AuthController struct {
	client *authSrc.AuthClient
	logger *logrus.Logger
}

// NewAuthController function.
// This function is used to create a new Auth controller.
//
// Parameters:
//   - registry (*consulDiscovery.ConsulRegistry): The Consul registry.
//   - logger (*logrus.Logger): The logger.
//
// Returns:
//   - *AuthController: The Auth controller.
func NewAuthController(registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) *AuthController {
	authServiceAddress, err := registry.GetServiceAddress(AUTH_SERVICE_NAME)
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Get Auth Service Address!")
	}

	authServiceClient, err := authSrc.NewAuthClient(authServiceAddress, "gateway-service")
	if err != nil {
		logger.WithFields(logrus.Fields{"error": err}).Fatal("Failed to Create Auth Service Client!")
	}

	return &AuthController{client: authServiceClient, logger: logger}
}

func (ac *AuthController) RegisterCompany(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.RegisterCompany")
	defer span.End()

	var req auth_models.CompanyRegisterPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Register Company!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.CompanyRegisterErrorResponse{
			Message: "Failed to Register Company!",
			Error:   err.Error(),
		})
		return
	}

	// validate req struct
	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Register Company!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.CompanyRegisterErrorResponse{
			Message: "Failed to Register Company!",
			Error:   err.Error(),
		})
		return
	}

	res, error := ac.client.RegisterCompany(ctx, req.Name, req.Description)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Register Customer!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.UserRegisterErrorResponse{
			Message: "Failed to Register Customer!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Company Registered Successfully!")
	c.JSON(http.StatusOK, auth_models.CompanyRegisterSuccessResponse{
		Message: "Company Registered Successfully",
		Company: auth_models.Company{
			ID:          res.Id,
			Name:        res.Name,
			Description: res.Description,
			CreatedAt:   res.CreatedAt.AsTime(),
			UpdatedAt:   res.UpdatedAt.AsTime(),
		},
	})
}

// ResendActivationEmail 		godoc
// @Summary 					Resend Activation Email
// @Description 				Resend Activation Email to a User
// @Tags 						Auth
// @Accept 						json
// @Produce 					json
// @Param 						payload body auth_models.ResendActivationEmailPayload true "Resend Activation Email Payload"
// @Success 					200 {object} auth_models.ResendActivationEmailSuccessResponse "Successfully resent activation email"
// @Failure 					400 {object} auth_models.ResendActivationEmailErrorResponse "Failed to resend activation email"
// @Failure 					500 {object} auth_models.ResendActivationEmailErrorResponse "Failed to resend activation email"
// @Router 					/auth/resend-activation-email [post]
func (ac *AuthController) ResendActivationEmail(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.ResendActivationEmail")
	defer span.End()

	var req auth_models.ResendActivationEmailPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Resend Activation Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.ResendActivationEmailErrorResponse{
			Message: "Failed to Resend Activation Email!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Resend Activation Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.ResendActivationEmailErrorResponse{
			Message: "Failed to Resend Activation Email!",
			Error:   err.Error(),
		})
		return
	}

	res, error := ac.client.ResendActivation(ctx, req.Identifier)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Resend Activation Email!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ResendActivationEmailErrorResponse{
			Message: "Failed to Resend Activation Email!",
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Activation Email Resent Successfully!")
	c.JSON(http.StatusOK, auth_models.ResendActivationEmailSuccessResponse{
		Message: "Activation Email Resent Successfully!",
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

// ForgotPassword 		godoc
// @Summary 			Forgot Password
// @Description 		Forgot Password for a User
// @Tags 				Auth
// @Accept 				json
// @Produce 			json
// @Param 				payload body auth_models.ForgotPasswordPayload true "Forgot Password Payload"
// @Success 			200 {object} auth_models.ForgotPasswordSuccessResponse "Successfully sent password reset email"
// @Failure 			400 {object} auth_models.ForgotPasswordErrorResponse "Failed to send password reset email"
// @Failure 			500 {object} auth_models.ForgotPasswordErrorResponse "Failed to send password reset email"
// @Router 				/auth/forgot-password [post]
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.ForgotPassword")
	defer span.End()

	var req auth_models.ForgotPasswordPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Send Password Reset Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.ForgotPasswordErrorResponse{
			Message: "Failed to Send Password Reset Email!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Send Password Reset Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.ForgotPasswordErrorResponse{
			Message: "Failed to Send Password Reset Email!",
			Error:   err.Error(),
		})
		return
	}

	err := ac.client.ForgotPassword(ctx, req.Identifier)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Send Password Reset Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ForgotPasswordErrorResponse{
			Message: "Failed to Send Password Reset Email!",
			Error:   err.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Password Reset Email Sent Successfully!")
	c.JSON(http.StatusOK, auth_models.ForgotPasswordSuccessResponse{
		Message: "Password Reset Email Sent Successfully!",
	})
}

// ResetPassword 		godoc
// @Summary 			Reset Password
// @Description 		Reset Password for a User
// @Tags 				Auth
// @Accept 				json
// @Produce 			json
// @Param 				passwordResetToken path string true "Password Reset Token"
// @Param 				payload body models.ResetPasswordPayload true "Reset Password Payload"
// @Success 			200 {object} models.ResetPasswordSuccessResponse "Successfully reset password"
// @Failure 			400 {object} models.ResetPasswordErrorResponse "Failed to reset password"
// @Failure 			500 {object} models.ResetPasswordErrorResponse "Failed to reset password"
// @Router 				/auth/reset-password/{passwordResetToken} [post]
func (ac *AuthController) ResetPassword(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.ResetPassword")
	defer span.End()

	passwordResetToken := c.Param("passwordResetToken")

	var req auth_models.ResetPasswordPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Reset Password!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.ResetPasswordErrorResponse{
			Message: "Failed to Reset Password!",
			Error:   err.Error(),
		})
		return
	}

	if err := Validate.Struct(req); err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Reset Password!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusBadRequest, auth_models.ResetPasswordErrorResponse{
			Message: "Failed to Reset Password!",
			Error:   err.Error(),
		})
		return
	}

	hashedPassword, err := password.HashPassword(req.Password, BCRYPT_HASHING_COST)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Reset Password!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ResetPasswordErrorResponse{
			Message: "Failed to Reset Password!",
			Error:   err.Error(),
		})
		return
	}

	err = ac.client.ResetPassword(ctx, passwordResetToken, hashedPassword)
	if err != nil {
		ac.logger.WithFields(logrus.Fields{"error": err}).Error("Failed to Reset Password!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ResetPasswordErrorResponse{
			Message: "Failed to Reset Password!",
			Error:   err.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Password Reset Successfully!")
	c.JSON(http.StatusOK, auth_models.ResetPasswordSuccessResponse{
		Message: "Password Reset Successfully!",
	})
}

func (ac *AuthController) GetCompanyByID(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GetCompanyByID")
	defer span.End()

	CompanyID := c.Param("companyId")
	if CompanyID == "" {
		ac.logger.WithFields(logrus.Fields{"error": "Company ID is required!"}).Error("Failed to Get Company by ID!")

		span.RecordError(errors.New("Company ID is required!"))
		span.SetStatus(otlpcodes.Error, "Company ID is required!")

		c.JSON(http.StatusBadRequest, auth_models.ErrorResponse{
			Message: "Failed to Get Company by ID!",
			Error:   "Company ID is required!",
		})
		return
	}

	res, error := ac.client.GetCompanyByID(ctx, CompanyID)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Get Company by ID!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{
			Message: "Failed to Get Company by ID!: " + CompanyID,
			Error:   error.Error(),
		})
		return
	}

	span.SetStatus(otlpcodes.Ok, "Company by ID Retrieved Successfully!")
	c.JSON(http.StatusOK, auth_models.Company{
		ID:          res.Id,
		Name:        res.Name,
		Description: res.Description,
		CreatedAt:   res.CreatedAt.AsTime(),
		UpdatedAt:   res.UpdatedAt.AsTime(),
	})
}

func (ac *AuthController) GetCompanies(c *gin.Context) {
	tracer := otel.Tracer("gateway-service")

	ctx := c.Request.Context()
	ctx, span := tracer.Start(ctx, "gatewayHTTPServer.GetCompanies")
	defer span.End()

	res, error := ac.client.GetAllCompany(ctx)
	if error != nil {
		ac.logger.WithFields(logrus.Fields{"error": error}).Error("Failed to Get Companies!")

		span.RecordError(error)
		span.SetStatus(otlpcodes.Error, error.Error())

		c.JSON(http.StatusInternalServerError, auth_models.ErrorResponse{
			Message: "Failed to Get Companies!",
			Error:   error.Error(),
		})
		return
	}

	companies := []auth_models.Company{}
	for _, company := range res.Companies {
		companies = append(companies, auth_models.Company{
			ID:          company.Id,
			Name:        company.Name,
			Description: company.Description,
			CreatedAt:   company.CreatedAt.AsTime(),
			UpdatedAt:   company.UpdatedAt.AsTime(),
		})
	}

	span.SetStatus(otlpcodes.Ok, "Companies Retrieved Successfully!")
	c.JSON(http.StatusOK, companies)
}
