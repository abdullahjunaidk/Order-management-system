package routes

import (
	consulDiscovery "common/discovery/consul"
	"gateway/controllers/auth_controller"
	"gateway/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthRoutes function.
// This function is used to define the auth routes.
//
// Parameters:
//   - router (*gin.RouterGroup): The router group.
//   - registry (*consulDiscovery.ConsulRegistry): The Consul registry.
//   - logger (*logrus.Logger): The logger.
//
// Routes:
//   - POST /auth/register: Register a user.
//   - GET /auth/activate/:activationToken: Activate a user.
//   - POST /auth/resend-activation-email: Resend activation email.
//   - POST /auth/forgot-password: Forgot password.
//   - POST /auth/reset-password/:passwordResetToken: Reset password.
//   - POST /auth/login: Login a user.
//   - POST /auth/refresh-access-token: Refresh access token.
//   - POST /auth/logout: Logout a user.
func AuthRoutes(router *gin.RouterGroup, registry *consulDiscovery.ConsulRegistry, logger *logrus.Logger) {
	authController := auth_controller.NewAuthController(registry, logger)

	authRouter := router.Group("/auth")

	// Company
	authRouter.POST("/company/register", authController.RegisterCompany)
	authRouter.GET("/company/:companyId", authController.GetCompanyByID)
	authRouter.GET("/companies", authController.GetCompanies)

	// User
	authRouter.POST("/user/register", authController.RegisterUser)
	authRouter.POST("/user/login", authController.LoginUser)
	authRouter.POST("/user/logout", authController.LogoutUser)
	authRouter.GET("/user/activate/:activationToken", authController.ActivateUser)
	authRouter.POST("/resend-activation-email", authController.ResendActivationEmail)
	authRouter.POST("/forgot-password", authController.ForgotPassword)
	authRouter.POST("/reset-password/:passwordResetToken", authController.ResetPassword)
	authRouter.POST("/refresh-access-token", authController.RefreshAccessToken)

	// Admin
	authRouter.POST("/admin/login", authController.LoginAdmin)
	authRouter.POST("/admin/logout", authController.LogoutAdmin)

	// Role
	authAdmin := router.Group("/auth/admin")
	authAdmin.Use(middlewares.AuthMiddleware(registry, logger))
	authAdmin.POST("/role/register", authController.RegisterRole)
	authAdmin.GET("/role/:roleId", authController.GetRoleByID)
	authAdmin.GET("/roles", authController.GetRoles)
}
