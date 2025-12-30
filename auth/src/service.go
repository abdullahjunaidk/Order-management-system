package src

import (
	rabbitmqBroker "common/broker/rabbitmq"
	"common/helpers/env"
	"common/helpers/jwt"
	pass "common/helpers/password"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"
)

var (
	// Activation Token Configuration.
	ACTIVATION_TOKEN_SECRET    = env.GetEnv("ACTIVATION_TOKEN_SECRET", "227b27db7136d2e82233db655e1cd19d9475bb8d8b7a3933f96ff5f526f1df64")
	ACTIVATION_TOKEN_EXPIRY, _ = time.ParseDuration(env.GetEnv("ACTIVATION_TOKEN_EXPIRY", "30m"))

	// Password Reset Token Configuration.
	PASSWORD_RESET_TOKEN_SECRET    = env.GetEnv("PASSWORD_RESET_TOKEN_SECRET", "75202ace94f6da7aca6d75e8ac22c727bf50e8d1703a88feb77df0993ea25167")
	PASSWORD_RESET_TOKEN_EXPIRY, _ = time.ParseDuration(env.GetEnv("PASSWORD_RESET_TOKEN_EXPIRY", "30m"))

	// Access Token Configuration.
	ACCESS_TOKEN_SECRET    = env.GetEnv("ACCESS_TOKEN_SECRET", "bfd0cd1c4b633a809f77ed67aa264f5a2a78b7e119ad52e44b9e3f2487038f14")
	ACCESS_TOKEN_EXPIRY, _ = time.ParseDuration(env.GetEnv("ACCESS_TOKEN_EXPIRY", "24h"))

	// Refresh Token Configuration.
	REFRESH_TOKEN_SECRET    = env.GetEnv("REFRESH_TOKEN_SECRET", "fd2c569d4dfb9892e3e0eca4314fa16ade9cf7ed04893c6568dcd409cde0a000")
	REFRESH_TOKEN_EXPIRY, _ = time.ParseDuration(env.GetEnv("REFRESH_TOKEN_EXPIRY", "72h"))

	// RabbitMQ Queues.
	USER_REGISTERED_QUEUE      = env.GetEnv("USER_REGISTERED_QUEUE", "auth.user.registered")
	USER_ACTIVATED_QUEUE       = env.GetEnv("USER_ACTIVATED_QUEUE", "auth.user.activated")
	USER_FORGOT_PASSWORD_QUEUE = env.GetEnv("USER_FORGOT_PASSWORD_QUEUE", "auth.user.forgotPassword")
	USER_PASSWORD_RESET_QUEUE  = env.GetEnv("USER_PASSWORD_RESET_QUEUE", "auth.user.passwordReset")
	EMPLOYEE_ACTIVATED_QUEUE   = env.GetEnv("EMPLOYEE_ACTIVATED_QUEUE", "auth.employee.activated")
)

// AuthService interface.
// This interface is used to define the auth service methods.
//
// Methods:
//   - RegisterBrand(ctx context.Context, payload BrandRegisterPayload) (*Brand, error): This method is user to register a new Brand
//   - RegisterUser(ctx context.Context, payload UserRegisterPayload) (*User, error): This method is used to register a new user.
//   - GetUserByID(ctx context.Context, userID string) (*User, error): This method is used to get the user by ID.
//   - GetUserByIdentifier(ctx context.Context, identifier string) (*User, error): This method is used to get the user by identifier.
//   - ActivateUser(ctx context.Context, userID string) (*User, error): This method is used to activate the user.
//   - ResendActivation(ctx context.Context, identifier string) (*User, error): This method is used to resend the activation email.
//   - ForgotPassword(ctx context.Context, identifier string) error: This method is used to send the password reset email.
//   - ResetPassword(ctx context.Context, passwordResetToken string, newPasswordHash string) error: This method is used to reset the password.
//   - LoginUser(ctx context.Context, identifier string) (*UserLoginResponse, error): This method is used to login the user.
//   - LogoutUser(ctx context.Context, identifier string) error: This method is used to logout the user.
//   - VerifyAccessToken(ctx context.Context, accessToken string) (*User, error): This method is used to verify the access token.
//   - RefreshAccessToken(ctx context.Context, refreshToken string) (*UserLoginResponse, error): This method is used to refresh the access token.
type AuthService interface {
	// Company
	RegisterCompany(ctx context.Context, name, description string) (*Company, error)
	GetCompanyByID(ctx context.Context, companyID string) (*Company, error)
	GetAllCompany(ctx context.Context) ([]Company, error)
	UpdateCompany(ctx context.Context, companyID, name, description string) (*Company, error)
	DeleteCompany(ctx context.Context, companyID string) error

	// User
	RegisterUser(ctx context.Context, payload RegisterUserPayload) (*User, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*User, error)
	ForgotPassword(ctx context.Context, identifier string) error
	ResetPassword(ctx context.Context, passwordResetToken string, newPasswordHash string) error
	LoginUser(ctx context.Context, identifier, password string) (*UserLoginResponse, error)
	LogoutUser(ctx context.Context, identifier string) error

	// Role
	RegisterRole(ctx context.Context, name, description string) (*Role, error)
	GetRoleByID(ctx context.Context, roleID string) (*Role, error)
	GetRoleByIdentifier(ctx context.Context, identifier string) (*Role, error)
	GetAllRole(ctx context.Context) ([]Role, error)

	// Token
	VerifyAccessToken(ctx context.Context, accessToken string) (*User, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*UserLoginResponse, error)
}

// authService struct.
// This struct is used to implement the AuthService interface.
//
// Attributes:
//   - store (AuthStore): The auth store.
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
type authService struct {
	store    AuthStore
	rabbitMQ *rabbitmqBroker.RabbitMQAdapter
}

// NewAuthService function.
// This function is used to create a new auth service.
//
// Parameters:
//   - store (AuthStore): The auth store.
//   - rabbitMQ (*rabbitmqBroker.RabbitMQAdapter): The RabbitMQ adapter.
//
// Returns:
//   - AuthService: The auth service.
func NewAuthService(store AuthStore, rabbitMQ *rabbitmqBroker.RabbitMQAdapter) AuthService {
	return &authService{
		store:    store,
		rabbitMQ: rabbitMQ,
	}
}

// RegisterBrand method.
// This method is used to register a new brand.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - payload (BrandRegisterPayload): The brand register payload.
//
// Returns:
//   - (*Brand, error): The brand and error.
func (s *authService) RegisterCompany(ctx context.Context, name, description string) (*Company, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.RegisterCompany")
	defer span.End()

	company := &Company{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Description: description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	createdCompanyID, err := s.store.RegisterCompany(ctx, company)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	createdCompany, err := s.store.GetCompanyByID(ctx, createdCompanyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Company Registered Successfully")
	return createdCompany, nil
}

func (s *authService) GetCompanyByID(ctx context.Context, companyID string) (*Company, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.GetCompanyByID")
	defer span.End()

	company, err := s.store.GetCompanyByID(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Company Found Successfully!")
	return company, nil
}

func (s *authService) GetAllCompany(ctx context.Context) ([]Company, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.GetAllCompany")
	defer span.End()

	companies, err := s.store.GetAllCompany(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Companies Found Successfully!")
	return companies, nil
}

func (s *authService) UpdateCompany(ctx context.Context, companyID string, name, description string) (*Company, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.UpdateCompany")
	defer span.End()

	company, err := s.store.GetCompanyByID(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	company.Name = name
	company.Description = description
	company.UpdatedAt = time.Now()

	updateErr := s.store.UpdateCompany(ctx, companyID, company)
	if updateErr != nil {
		span.RecordError(updateErr)
		span.SetStatus(otlpcodes.Error, updateErr.Error())
		return nil, updateErr
	}

	updatedCompany, err := s.store.GetCompanyByID(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Company Updated Successfully")
	return updatedCompany, nil
}

func (s *authService) DeleteCompany(ctx context.Context, companyID string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.DeleteCompany")
	defer span.End()

	err := s.store.DeleteCompany(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Company Deleted Successfully")
	return nil
}

// RegisterUser method.
// This method is used to register a new user.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - payload (UserRegisterPayload): The user register payload.
//
// Returns:
//   - (*User, error): The user and error.
func (s *authService) RegisterUser(ctx context.Context, payload RegisterUserPayload) (*User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.RegisterUser")
	defer span.End()

	user := &User{

		ID:           primitive.NewObjectID(),
		Name:         payload.Name,
		Username:     payload.Username,
		Email:        payload.Email,
		PasswordHash: payload.PasswordHash,
		Phone:        payload.Phone,
		Incentive:    0,
		IsActive:     true,
		IsSuperAdmin: false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdUserID, err := s.store.RegisterUser(ctx, user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	createdUser, err := s.store.GetUserByID(ctx, createdUserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Registered Successfully")
	return createdUser, nil
}

func (s *authService) GetUserByIdentifier(ctx context.Context, identifier string) (*User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.GetUserByIdentifier")
	defer span.End()

	user, err := s.store.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Found Successfully!")
	return user, nil
}

// GetUserByID method.
// This method is used to get the user by ID.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - userID (string): The user ID.
//
// Returns:
//   - *User: The user.
//   - error: The error.
func (s *authService) GetUserByID(ctx context.Context, userID string) (*User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.GetUserByID")
	defer span.End()

	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Found Successfully!")
	return user, nil
}

// ForgotPassword method.
// This method is used to send the password reset email.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - identifier (string): The identifier.
//
// Returns:
//   - error: The error.
func (s *authService) ForgotPassword(ctx context.Context, identifier string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.ForgotPassword")
	defer span.End()

	user, err := s.store.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	passwordResetToken, err := jwt.GeneratePasswordResetToken(user.ID.Hex(), user.IsSuperAdmin, user.CompanyIDs, PASSWORD_RESET_TOKEN_EXPIRY, PASSWORD_RESET_TOKEN_SECRET)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	user.PasswordResetToken = passwordResetToken

	if err := s.store.UpdateUser(ctx, user.ID.Hex(), user); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Password Reset Email Sent Successfully!")
	return nil
}

// ResetPassword method.
// This method is used to reset the password.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - passwordResetToken (string): The password reset token.
//   - newPasswordHash (string): The new password hash.
//
// Returns:
//   - error: The error.
func (s *authService) ResetPassword(ctx context.Context, passwordResetToken string, newPasswordHash string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.ResetPassword")
	defer span.End()

	claims, err := jwt.VerifyPasswordResetToken(passwordResetToken, PASSWORD_RESET_TOKEN_SECRET)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	user, err := s.store.GetUserByID(ctx, claims["user_id"].(string))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	if passwordResetToken != user.PasswordResetToken {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, "Invalid Password Reset Token!")
		return errors.New("Invalid Password Reset Token!")
	}

	user.PasswordHash = newPasswordHash
	user.PasswordResetToken = ""
	user.UpdatedAt = time.Now()

	if err := s.store.UpdateUser(ctx, user.ID.Hex(), user); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "Password Reset Successfully!")
	return nil
}

func (s *authService) LoginUser(ctx context.Context, identifier, password string) (*UserLoginResponse, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.LoginUser")
	defer span.End()

	user, err := s.store.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	if !user.IsActive {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, "User is not active!")
		return nil, errors.New("User is not active!")
	}

	if err := pass.ComparePassword(user.PasswordHash, password); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, "Invalid Password!")
		return nil, errors.New("Invalid Password!")
	}

	accessToken := user.AccessToken
	refreshToken := user.RefreshToken

	if accessToken == "" || refreshToken == "" {
		accessToken, err = jwt.GenerateAccessToken(user.ID.Hex(), user.IsSuperAdmin, user.CompanyIDs, ACCESS_TOKEN_EXPIRY, ACCESS_TOKEN_SECRET)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return nil, err
		}

		refreshToken, err = jwt.GenerateRefreshToken(user.ID.Hex(), user.IsSuperAdmin, user.CompanyIDs, REFRESH_TOKEN_EXPIRY, REFRESH_TOKEN_SECRET)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			return nil, err
		}
	} else {
		_, accessErr := jwt.VerifyAccessToken(accessToken, ACCESS_TOKEN_SECRET)
		if accessErr != nil {
			_, refreshErr := jwt.VerifyRefreshToken(refreshToken, REFRESH_TOKEN_SECRET)
			if refreshErr != nil {
				accessToken, err = jwt.GenerateAccessToken(user.ID.Hex(), user.IsSuperAdmin, user.CompanyIDs, ACCESS_TOKEN_EXPIRY, ACCESS_TOKEN_SECRET)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(otlpcodes.Error, err.Error())
					return nil, err
				}

				refreshToken, err = jwt.GenerateRefreshToken(user.ID.Hex(), user.IsSuperAdmin, user.CompanyIDs, REFRESH_TOKEN_EXPIRY, REFRESH_TOKEN_SECRET)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(otlpcodes.Error, err.Error())
					return nil, err
				}
			}
		}
	}

	user.AccessToken = accessToken
	user.RefreshToken = refreshToken

	if err := s.store.UpdateUser(ctx, user.ID.Hex(), user); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "User Logged In Successfully!")
	return &UserLoginResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// LogoutEmployee method.
// This method is used to logout the user.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - identifier (string): The user identifier.
//
// Returns:
//   - error: The error.
func (s *authService) LogoutUser(ctx context.Context, identifier string) error {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.LogoutUser")
	defer span.End()

	user, err := s.store.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	user.AccessToken = ""
	user.RefreshToken = ""
	if err := s.store.UpdateUser(ctx, user.ID.Hex(), user); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return err
	}

	span.SetStatus(otlpcodes.Ok, "User Logged Out Successfully!")
	return nil
}

// VerifyAccessToken method.
// This method is used to verify the access token.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - accessToken (string): The access token.
//
// Returns:
//   - *Employee: The employee.
//   - error: The error.
func (s *authService) VerifyAccessToken(ctx context.Context, accessToken string) (*User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.VerifyAccessToken")
	defer span.End()

	claims, err := jwt.VerifyAccessToken(accessToken, ACCESS_TOKEN_SECRET)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	user, err := s.store.GetUserByID(ctx, claims["userId"].(string))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	if accessToken != user.AccessToken {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, "Invalid Access Token!")
		return nil, errors.New("invalid or expired access token")
	}

	span.SetStatus(otlpcodes.Ok, "Access Token Verified Successfully!")
	return user, nil
}

// RefreshAccessToken method.
// This method is used to refresh the access token.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - refreshToken (string): The refresh token.
//
// Returns:
//   - *UserLoginResponse: The user login response.
//   - error: The error.
func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (*UserLoginResponse, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.RefreshAccessToken")
	defer span.End()

	claims, err := jwt.VerifyRefreshToken(refreshToken, REFRESH_TOKEN_SECRET)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	user, err := s.store.GetUserByID(ctx, claims["userId"].(string))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	if refreshToken != user.RefreshToken {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, "Invalid Refresh Token!")
		return nil, errors.New("invalid or expired refresh token")
	}

	accessToken, err := jwt.GenerateAccessToken(user.ID.Hex(), user.IsSuperAdmin, user.CompanyIDs, ACCESS_TOKEN_EXPIRY, ACCESS_TOKEN_SECRET)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	user.AccessToken = accessToken
	if err := s.store.UpdateUser(ctx, user.ID.Hex(), user); err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	userLoginResponse := &UserLoginResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	span.SetStatus(otlpcodes.Ok, "Access Token Refreshed Successfully!")
	return userLoginResponse, nil
}

func (s *authService) RegisterRole(ctx context.Context, name, description string) (*Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.RegisterRole")
	defer span.End()

	role := &Role{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_, err := s.store.RegisterRole(ctx, role)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Role Registered Successfully!")
	return role, nil
}

func (s *authService) GetRoleByID(ctx context.Context, roleID string) (*Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.GetRoleByID")
	defer span.End()

	role, err := s.store.GetRoleByID(ctx, roleID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Role Found Successfully!")
	return role, nil
}

func (s *authService) GetRoleByIdentifier(ctx context.Context, identifier string) (*Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.GetRoleByIdentifier")
	defer span.End()

	role, err := s.store.GetRoleByIdentifier(ctx, identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Role Found Successfully!")
	return role, nil
}

func (s *authService) GetAllRole(ctx context.Context) ([]Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authService.GetAllRoles")
	defer span.End()

	roles, err := s.store.GetAllRoles(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(otlpcodes.Ok, "Roles Found Successfully!")
	return roles, nil
}

// // ActivateUser method.
// // This method is used to activate the user.
// //
// // Parameters:
// //   - ctx (context.Context): The context.
// //   - activationToken (string): The activation token.
// //
// // Returns:
// //   - *User: The user.
// //   - error: The error.
// func (s *authService) ActivateUser(ctx context.Context, activationToken string) (*User, error) {
// 	tracer := otel.Tracer("auth-service")
// 	ctx, span := tracer.Start(ctx, "authService.ActivateUser")
// 	defer span.End()

// 	claims, err := jwt.VerifyActivationToken(activationToken, ACTIVATION_TOKEN_SECRET)
// 	if err != nil {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())
// 		return nil, err
// 	}

// 	user, err := s.store.GetUserByID(ctx, claims["user_id"].(string))
// 	if err != nil {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())
// 		return nil, err
// 	}

// 	if activationToken != user.ActivationToken {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, "Invalid Activation Token!")
// 		return nil, errors.New("Invalid Activation Token!")
// 	}

// 	user.IsActive = true
// 	user.ActivationToken = ""
// 	user.UpdatedAt = time.Now()

// 	if err := s.store.UpdateUser(ctx, user.ID.Hex(), user); err != nil {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())
// 		return nil, err
// 	}

// 	eventPayload := UserActivatedEventPayload{
// 		Username: user.Username,
// 		Email:    user.Email,
// 	}

// 	if err := s.rabbitMQ.PublishMessage(ctx, USER_ACTIVATED_QUEUE, eventPayload); err != nil {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())
// 		return nil, err
// 	}

// 	span.SetStatus(otlpcodes.Ok, "User Activated Successfully!")
// 	return user, nil
// }

// // ResendActivation method.
// // This method is used to resend the activation email.
// //
// // Parameters:
// //   - ctx (context.Context): The context.
// //   - identifier (string): The identifier.
// //
// // Returns:
// //   - *User: The user.
// //   - error: The error.
// func (s *authService) ResendActivation(ctx context.Context, identifier string) (*User, error) {
// 	tracer := otel.Tracer("auth-service")
// 	ctx, span := tracer.Start(ctx, "authService.ResendActivation")
// 	defer span.End()

// 	user, err := s.store.GetUserByIdentifier(ctx, identifier)
// 	if err != nil {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())
// 		return nil, err
// 	}

// 	if user.IsActive {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, "User Already Activated!")
// 		return nil, errors.New("User Already Activated!")
// 	}

// 	activationToken, err := jwt.GenerateActivationToken(user.ID.Hex(), ACTIVATION_TOKEN_EXPIRY, ACTIVATION_TOKEN_SECRET)
// 	if err != nil {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())
// 		return nil, err
// 	}

// 	user.ActivationToken = activationToken
// 	user.PasswordResetToken = ""

// 	if err := s.store.UpdateUser(ctx, user.ID.Hex(), user); err != nil {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())
// 		return nil, err
// 	}

// 	eventPayload := UserRegisteredEventPayload{
// 		ActivationToken: activationToken,
// 		Username:        user.Username,
// 		Email:           user.Email,
// 	}

// 	if err := s.rabbitMQ.PublishMessage(ctx, USER_REGISTERED_QUEUE, eventPayload); err != nil {
// 		span.RecordError(err)
// 		span.SetStatus(otlpcodes.Error, err.Error())
// 		return nil, err
// 	}

// 	span.SetStatus(otlpcodes.Ok, "Activation Email Resent Successfully!")
// 	return user, nil
// }
