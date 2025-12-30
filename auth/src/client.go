package src

import (
	"common/helpers/logger"
	authProto "common/proto/auth"
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AuthClient struct.
// This struct is used to create a new client connection to the auth service.
//
// Attributes:
//   - conn (*grpc.ClientConn): The client connection to the auth service.
//   - service (authProto.AuthServiceClient): The auth service client.
//   - tracerName (string): The tracer name.
//   - log (*logrus.Logger): The logger.
type AuthClient struct {
	conn       *grpc.ClientConn
	service    authProto.AuthServiceClient
	tracerName string
	log        *logrus.Logger
}

// --- Client Lifecycle ---

// NewAuthClient function.
// This function is used to create a new client connection to the auth service.
//
// Parameters:
//   - address (string): The address of the auth service.
//
// Returns:
//   - *AuthClient: The auth client.
//   - error: The error.
func NewAuthClient(address string, tracerName string) (*AuthClient, error) {
	log := logger.NewLogger()

	clientOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.NewClient(address, clientOpts...)
	if err != nil {
		log.WithFields(logrus.Fields{"address": address, "error": err}).Error("Failed to Create Client Connection!")
		return nil, fmt.Errorf("failed to create client connection: %w", err)
	}

	service := authProto.NewAuthServiceClient(conn)

	return &AuthClient{
		conn:       conn,
		service:    service,
		tracerName: tracerName,
		log:        log,
	}, nil
}

// Close function.
// This function is used to close the client connection to the auth service.
func (c *AuthClient) Close() {
	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			c.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Close Connection!")
		} else {
			c.log.Info("Connection Closed!")
		}
	}
}

// --- Company Management ---

func (c *AuthClient) RegisterCompany(ctx context.Context, name, description string) (*authProto.Company, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.RegisterCompany")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.RegisterCompany(ctx, &authProto.CompanyRegisterPayload{
		Name:        name,
		Description: description,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"name": name, "description": description, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Register Company!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("Register Company deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("Failed to Register Company: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"name": name, "description": description, "error": err}).Error("Failed to Register Company!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("Failed to Register Company: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Company Registered Successfully via gRPC")
	return res, nil
}

func (c *AuthClient) GetCompanyByID(ctx context.Context, companyID string) (*authProto.Company, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.GetCompanyByID")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.GetCompanyByID(ctx, &authProto.GetCompanyByIDPayload{Id: companyID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"company_id": companyID, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Get Company By ID!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("get company by ID deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to get company by ID: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"company_id": companyID, "error": err}).Error("Failed to Get Company By ID!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to get company by ID: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Company Found Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) GetCompanyByIdentifier(ctx context.Context, identifier string) (*authProto.Company, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.GetCompanyByIdentifier")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.GetCompanyByIdentifier(ctx, &authProto.GetCompanyByIdentifierPayload{Identifier: identifier})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Get Company By Identifier!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("get company by identifier deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to get company by identifier: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err}).Error("Failed to Get Company By Identifier!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to get company by identifier: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Company Found Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) GetAllCompany(ctx context.Context) (*authProto.CompanyList, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.GetAllCompany")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.GetAllCompany(ctx, &emptypb.Empty{})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Get All Company!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("get all company deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to get all company: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Get All Company!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to get all company: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "All Companies Found Successfully Via GRPC!")
	return res, nil
}

// --- User Management ---

func (c *AuthClient) RegisterUser(ctx context.Context, name, username, email, passwordHash string, phone int64, companyIds []string) (*authProto.User, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.RegisterUser")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.RegisterUser(ctx, &authProto.UserRegisterPayload{
		Name:         name,
		Username:     username,
		Email:        email,
		Phone:        phone,
		PasswordHash: passwordHash,
		CompanyIds:   companyIds,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"name": name, "username": username, "email": email, "phone": phone, "company_ids": companyIds, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Register User!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("Register User deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("Failed to Register User: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"name": name, "username": username, "email": email, "phone": phone, "company_ids": companyIds, "error": err}).Error("Failed to Register User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("Failed to Register User: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Registered Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) GetUserByID(ctx context.Context, userID string) (*authProto.User, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.GetUserByID")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.GetUserById(ctx, &authProto.GetUserByIdPayload{Id: userID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"user_id": userID, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Get User By ID!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("get user by ID deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to get user by ID: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"user_id": userID, "error": err}).Error("Failed to Get User By ID!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Retrieved Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) GetUserByIdentifier(ctx context.Context, identifier string) (*authProto.User, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.GetUserByIdentifier")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.GetUserByIdentifier(ctx, &authProto.GetUserByIdentifierPayload{Identifier: identifier})

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Get User!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("Get User deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("Failed to Get User: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err}).Error("Failed to Get User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("Failed to Get User: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Found Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) ActivateUser(ctx context.Context, activationToken string) (*authProto.User, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.ActivateUser")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.ActivateUser(ctx, &authProto.ActivateUserPayload{ActivationToken: activationToken})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"activation_token": activationToken, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Activate User!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("activate user deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to activate user: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"activation_token": activationToken, "error": err}).Error("Failed to Activate User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Activated Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) ResendActivation(ctx context.Context, identifier string) (*authProto.User, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.ResendActivation")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.ResendActivation(ctx, &authProto.ResendActivationPayload{Identifier: identifier})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Resend Activation!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("resend activation deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to resend activation: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err}).Error("Failed to Resend Activation!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to resend activation: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Resend Activation Email Sent Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) ForgotPassword(ctx context.Context, identifier string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.ForgotPassword")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.service.ForgotPassword(ctx, &authProto.ForgotPasswordPayload{Identifier: identifier})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Send Forgot Password Email!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return fmt.Errorf("forgot password deadline exceeded: %w", err)
			}

			return fmt.Errorf("failed to send forgot password email: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err}).Error("Failed to Send Forgot Password Email!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return fmt.Errorf("failed to send forgot password email: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Forgot Password Email Sent Successfully Via GRPC!")
	return nil
}

func (c *AuthClient) ResetPassword(ctx context.Context, passwordResetToken, newPasswordHash string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.ResetPassword")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.service.ResetPassword(ctx, &authProto.ResetPasswordPayload{PasswordResetToken: passwordResetToken, PasswordHash: newPasswordHash})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"password_reset_token": passwordResetToken, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Reset Password!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return fmt.Errorf("reset password deadline exceeded: %w", err)
			}

			return fmt.Errorf("failed to reset password: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"password_reset_token": passwordResetToken, "error": err}).Error("Failed to Reset Password!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return fmt.Errorf("failed to reset password: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Password Reset Successfully Via GRPC!")
	return nil
}

func (c *AuthClient) LoginUser(ctx context.Context, identifier string, passwordHash string) (*authProto.LoginUserResponse, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.LoginUser")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.LoginUser(ctx, &authProto.LoginUserPayload{Identifier: identifier, PasswordHash: passwordHash})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Login User!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("login user deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to login user: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err}).Error("Failed to Login User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to login user: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Logged In Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) LogoutUser(ctx context.Context, identifier string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.LogoutUser")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.service.LogoutUser(ctx, &authProto.LogoutUserPayload{Identifier: identifier})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Logout User!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return fmt.Errorf("logout user deadline exceeded: %w", err)
			}

			return fmt.Errorf("failed to logout user: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Logout User!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return fmt.Errorf("failed to logout user: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Logged Out Successfully Via GRPC!")
	return nil
}

// --- Admin Management ---

func (c *AuthClient) LoginAdmin(ctx context.Context, identifier, password string) (*authProto.LoginUserResponse, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.LoginAdmin")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.LoginUser(ctx, &authProto.LoginUserPayload{Identifier: identifier, PasswordHash: password})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"username": identifier, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Login Admin!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("login admin deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to login admin: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"username": identifier, "error": err}).Error("Failed to Login Admin!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to login admin: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Admin Logged In Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) LogoutAdmin(ctx context.Context, identifier string) error {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.LogoutAdmin")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.service.LogoutUser(ctx, &authProto.LogoutUserPayload{Identifier: identifier})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Logout User!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return fmt.Errorf("logout admin deadline exceeded: %w", err)
			}

			return fmt.Errorf("failed to logout admin: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err}).Error("Failed to Logout Admin!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return fmt.Errorf("failed to logout admin: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Admin Logged Out Successfully Via GRPC!")
	return nil
}

// --- Token Management ---

func (c *AuthClient) VerifyAccessToken(ctx context.Context, accessToken string) (*authProto.User, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.VerifyAccessToken")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.VerifyAccessToken(ctx, &authProto.VerifyAccessTokenPayload{AccessToken: accessToken})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"access_token": accessToken, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Verify Access Token!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("verify access token deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to verify access token: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"access_token": accessToken, "error": err}).Error("Failed to Verify Access Token!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to verify access token: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Access Token Verified Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) RefreshAccessToken(ctx context.Context, refreshToken string) (*authProto.LoginUserResponse, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.RefreshAccessToken")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.RefreshAccessToken(ctx, &authProto.RefreshAccessTokenPayload{RefreshToken: refreshToken})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"refresh_token": refreshToken, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Refresh Access Token!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("refresh access token deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to refresh access token: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"refresh_token": refreshToken, "error": err}).Error("Failed to Refresh Access Token!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Access Token Refreshed Successfully Via GRPC!")
	return res, nil
}

// --- Role Management ---

func (c *AuthClient) RegisterRole(ctx context.Context, name string) (*authProto.Role, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.RegisterRole")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.RegisterRole(ctx, &authProto.RoleRegisterPayload{Name: name})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"name": name, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Register Role!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("register role deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to register role: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"name": name, "error": err}).Error("Failed to Register Role!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to register role: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role Registered Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) GetRoleByID(ctx context.Context, roleID string) (*authProto.Role, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.GetRoleByID")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.GetRoleById(ctx, &authProto.GetRoleByIdPayload{Id: roleID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"role_id": roleID, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Get Role By ID!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("get role by id deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to get role by id: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"role_id": roleID, "error": err}).Error("Failed to Get Role By ID!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to get role by id: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role By ID Retrieved Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) GetRoleByIdentifier(ctx context.Context, identifier string) (*authProto.Role, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.GetRoleByIdentifier")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.GetRoleByIdentifier(ctx, &authProto.GetRoleByIdentifierPayload{Identifier: identifier})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Get Role By Identifier!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("get role by identifier deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to get role by identifier: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"identifier": identifier, "error": err}).Error("Failed to Get Role By Identifier!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to get role by identifier: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role By Identifier Retrieved Successfully Via GRPC!")
	return res, nil
}

func (c *AuthClient) GetRoles(ctx context.Context) (*authProto.RoleList, error) {
	tracer := otel.Tracer(c.tracerName)
	ctx, span := tracer.Start(ctx, "authClient.GetRoles")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := c.service.GetRoles(ctx, &emptypb.Empty{})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			c.log.WithFields(logrus.Fields{"error": err, "grpc_code": st.Code(), "grpc_message": st.Message()}).Error("Failed to Get All Roles!")

			span.RecordError(err)
			span.SetStatus(otlpcodes.Error, err.Error())
			if st.Code() == codes.DeadlineExceeded {
				return nil, fmt.Errorf("get all roles deadline exceeded: %w", err)
			}

			return nil, fmt.Errorf("failed to get all roles: %s", st.Message())
		}
		c.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Get All Roles!")

		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}

	span.SetStatus(otlpcodes.Ok, "All Roles Retrieved Successfully Via GRPC!")
	return res, nil
}
