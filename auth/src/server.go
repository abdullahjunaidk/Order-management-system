package src

import (
	"common/helpers/logger"
	authProto "common/proto/auth"
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/otel"
	otlpcodes "go.opentelemetry.io/otel/codes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

// authGRPCServer struct.
// This struct is used to implement the auth service gRPC server.
//
// Attributes:
//   - authService (AuthService): The auth service.
//   - log (*logrus.Logger): The logger.
type authGRPCServer struct {
	authProto.UnimplementedAuthServiceServer
	authService AuthService
	log         *logrus.Logger
}

// --- Server Lifecycle ---

// ListenAndServeGRPC function.
// This function is used to listen and serve the auth service gRPC server.
//
// Parameters:
//   - service (AuthService): The auth service.
//   - port (int): The port.
//
// Returns:
//   - error: The error.
func ListenAndServeGRPC(service AuthService, port int) error {
	log := logger.NewLogger()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithFields(logrus.Fields{"port": port, "error": err}).Error("Failed to Listen!")
		return fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()

	authProto.RegisterAuthServiceServer(server, &authGRPCServer{authService: service, log: log})
	grpc_health_v1.RegisterHealthServer(server, &healthServer{})
	reflection.Register(server)

	log.WithFields(logrus.Fields{"port": port}).Info("Starting gRPC Server...")
	if err := server.Serve(lis); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Serve gRPC Server!")
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// --- Company Management ---

func (s *authGRPCServer) RegisterCompany(ctx context.Context, req *authProto.CompanyRegisterPayload) (*authProto.Company, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.RegisterCompany")
	defer span.End()

	s.log.WithFields(logrus.Fields{"name": req.Name, "description": req.Description}).Info("Registering New Company...")

	res, err := s.authService.RegisterCompany(ctx, req.Name, req.Description)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"name": req.Name, "description": req.Description, "error": err}).Error("Failed to Register Company!")
		return nil, status.Errorf(codes.Internal, "failed to register company: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Company Registered Succeessfully")
	return &authProto.Company{
		Id:          res.ID.Hex(),
		Name:        res.Name,
		Description: res.Description,
		CreatedAt:   timestamppb.New(res.CreatedAt),
		UpdatedAt:   timestamppb.New(res.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) GetCompanyByID(ctx context.Context, req *authProto.GetCompanyByIDPayload) (*authProto.Company, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetCompanyByID")
	defer span.End()

	company, err := s.authService.GetCompanyByID(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"company_id": req.Id, "error": err}).Error("Failed to Get Company!")
		return nil, status.Errorf(codes.Internal, "failed to get company: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Company Found Successfully!")
	return &authProto.Company{
		Id:          company.ID.Hex(),
		Name:        company.Name,
		Description: company.Description,
		CreatedAt:   timestamppb.New(company.CreatedAt),
		UpdatedAt:   timestamppb.New(company.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) GetAllCompany(ctx context.Context, req *emptypb.Empty) (*authProto.CompanyList, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetAllCompany")
	defer span.End()

	companies, err := s.authService.GetAllCompany(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Get All Companies!")
		return nil, status.Errorf(codes.Internal, "failed to get all companies: %v", err)
	}

	companiesProto := make([]*authProto.Company, len(companies))
	for i, company := range companies {
		companiesProto[i] = &authProto.Company{
			Id:          company.ID.Hex(),
			Name:        company.Name,
			Description: company.Description,
			CreatedAt:   timestamppb.New(company.CreatedAt),
			UpdatedAt:   timestamppb.New(company.UpdatedAt),
		}
	}

	span.SetStatus(otlpcodes.Ok, "All Companies Found Successfully!")
	return &authProto.CompanyList{Companies: companiesProto}, nil
}

// --- User Management ---

func (s *authGRPCServer) RegisterUser(ctx context.Context, req *authProto.UserRegisterPayload) (*authProto.User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.RegisterUser")
	defer span.End()

	payload := RegisterUserPayload{
		Name:         req.Name,
		Username:     req.Username,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: req.PasswordHash,
		CompanyIDs:   req.CompanyIds,
	}

	res, err := s.authService.RegisterUser(ctx, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"name": req.Name, "username": req.Username, "email": req.Email, "phone": req.Phone, "error": err}).Error("Failed to Register User!")
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Registered Successfully!")
	return &authProto.User{
		Id:           res.ID.Hex(),
		Name:         res.Name,
		Username:     res.Username,
		Email:        res.Email,
		Phone:        res.Phone,
		PasswordHash: res.PasswordHash,
		Incentive:    res.Incentive,
		IsActive:     res.IsActive,
		IsSuperAdmin: res.IsSuperAdmin,
		CreatedAt:    timestamppb.New(res.CreatedAt),
		UpdatedAt:    timestamppb.New(res.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) GetUserByIdentifier(ctx context.Context, req *authProto.GetUserByIdentifierPayload) (*authProto.User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetUserByIdentifier")
	defer span.End()

	res, err := s.authService.GetUserByIdentifier(ctx, req.Identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"identifier": req.Identifier, "error": err}).Error("Failed to Get User!")
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Found Successfully!")
	return &authProto.User{
		Id:           res.ID.Hex(),
		Name:         res.Name,
		Username:     res.Username,
		Email:        res.Email,
		Phone:        res.Phone,
		PasswordHash: res.PasswordHash,
		Incentive:    res.Incentive,
		IsActive:     res.IsActive,
		IsSuperAdmin: res.IsSuperAdmin,
		CreatedAt:    timestamppb.New(res.CreatedAt),
		UpdatedAt:    timestamppb.New(res.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) GetUserById(ctx context.Context, req *authProto.GetUserByIdPayload) (*authProto.User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetUserById")
	defer span.End()

	s.log.WithFields(logrus.Fields{"user_id": req.Id}).Info("Getting User By ID...")

	res, err := s.authService.GetUserByID(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"user_id": req.Id, "error": err}).Error("Failed to Get User By ID!")
		return nil, status.Errorf(codes.Internal, "failed to get user by ID: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Retrieved Successfully!")
	return &authProto.User{
		Id:           res.ID.Hex(),
		Name:         res.Name,
		Username:     res.Username,
		Email:        res.Email,
		Phone:        res.Phone,
		PasswordHash: res.PasswordHash,
		Incentive:    res.Incentive,
		IsActive:     res.IsActive,
		IsSuperAdmin: res.IsSuperAdmin,
		CreatedAt:    timestamppb.New(res.CreatedAt),
		UpdatedAt:    timestamppb.New(res.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) ForgotPassword(ctx context.Context, req *authProto.ForgotPasswordPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.ForgotPassword")
	defer span.End()

	s.log.WithFields(logrus.Fields{"identifier": req.Identifier}).Info("Forgot Password...")

	err := s.authService.ForgotPassword(ctx, req.Identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"identifier": req.Identifier, "error": err}).Error("Failed to Forgot Password!")
		return nil, status.Errorf(codes.Internal, "failed to forgot password: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Forgot Password Email Sent Successfully!")
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) ResetPassword(ctx context.Context, req *authProto.ResetPasswordPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.ResetPassword")
	defer span.End()

	s.log.WithFields(logrus.Fields{"reset_token": req.PasswordResetToken}).Info("Resetting Password...")

	err := s.authService.ResetPassword(ctx, req.PasswordResetToken, req.PasswordHash)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"reset_token": req.PasswordResetToken, "error": err}).Error("Failed to Reset Password!")
		return nil, status.Errorf(codes.Internal, "failed to reset password: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Password Reset Successfully!")
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) LoginUser(ctx context.Context, req *authProto.LoginUserPayload) (*authProto.LoginUserResponse, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.LoginUser")
	defer span.End()

	s.log.WithFields(logrus.Fields{"identifier": req.Identifier}).Info("Logging In User...")

	res, err := s.authService.LoginUser(ctx, req.Identifier, req.PasswordHash)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"identifier": req.Identifier, "error": err}).Error("Failed to Login User!")
		return nil, status.Errorf(codes.Internal, "failed to login employee: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Employee Logged In Successfully!")
	return &authProto.LoginUserResponse{
		User: &authProto.User{
			Id:           res.User.ID.Hex(),
			Name:         res.User.Name,
			Username:     res.User.Username,
			Email:        res.User.Email,
			Phone:        res.User.Phone,
			PasswordHash: res.User.PasswordHash,
			Incentive:    res.User.Incentive,
			IsActive:     res.User.IsActive,
			IsSuperAdmin: res.User.IsSuperAdmin,
			CreatedAt:    timestamppb.New(res.User.CreatedAt),
			UpdatedAt:    timestamppb.New(res.User.UpdatedAt),
		},
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}, nil
}

func (s *authGRPCServer) LogoutUser(ctx context.Context, req *authProto.LogoutUserPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.LogoutUser")
	defer span.End()

	s.log.WithFields(logrus.Fields{"identifier": req.Identifier}).Info("Logging Out User...")

	err := s.authService.LogoutUser(ctx, req.Identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"identifier": req.Identifier, "error": err}).Error("Failed to Logout User!")
		return nil, status.Errorf(codes.Internal, "failed to logout User: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "User Logged Out Successfully!")
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) GrantCompanyAccess(ctx context.Context, req *authProto.GrantCompanyAccessPayload) (*authProto.User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GrantCompanyAccess")
	defer span.End()

	s.log.WithFields(logrus.Fields{"user_id": req.UserId, "company_ids": req.CompanyIds}).Info("Granting Company Access to User...")

	payload := GrantCompanyAccessPayload{
		UserID:     req.UserId,
		CompanyIDs: req.CompanyIds,
	}

	res, err := s.authService.GrantCompanyAccess(ctx, payload)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"user_id": req.UserId, "error": err}).Error("Failed to Grant Company Access!")
		return nil, status.Errorf(codes.Internal, "failed to grant company access: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Company Access Granted Successfully!")
	return &authProto.User{
		Id:           res.ID.Hex(),
		Name:         res.Name,
		Username:     res.Username,
		Email:        res.Email,
		Phone:        res.Phone,
		PasswordHash: res.PasswordHash,
		Incentive:    res.Incentive,
		IsActive:     res.IsActive,
		IsSuperAdmin: res.IsSuperAdmin,
		CompanyIds:   res.CompanyIDs,
		CreatedAt:    timestamppb.New(res.CreatedAt),
		UpdatedAt:    timestamppb.New(res.UpdatedAt),
	}, nil
}

// --- Token Management ---

func (s *authGRPCServer) VerifyAccessToken(ctx context.Context, req *authProto.VerifyAccessTokenPayload) (*authProto.User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.VerifyAccessToken")
	defer span.End()

	s.log.WithFields(logrus.Fields{"access_token": req.AccessToken}).Info("Verifying Access Token...")

	res, err := s.authService.VerifyAccessToken(ctx, req.AccessToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"access_token": req.AccessToken, "error": err}).Error("Failed to Verify Access Token!")
		return nil, status.Errorf(codes.Internal, "failed to verify access token: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Access Token Verified Successfully!")
	return &authProto.User{
		Id:           res.ID.Hex(),
		Name:         res.Name,
		Username:     res.Username,
		Email:        res.Email,
		Phone:        res.Phone,
		PasswordHash: res.PasswordHash,
		Incentive:    res.Incentive,
		IsActive:     res.IsActive,
		IsSuperAdmin: res.IsSuperAdmin,
		CreatedAt:    timestamppb.New(res.CreatedAt),
		UpdatedAt:    timestamppb.New(res.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) RefreshAccessToken(ctx context.Context, req *authProto.RefreshAccessTokenPayload) (*authProto.LoginUserResponse, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.RefreshAccessToken")
	defer span.End()

	s.log.WithFields(logrus.Fields{"refresh_token": req.RefreshToken}).Info("Refreshing Access Token...")

	res, err := s.authService.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"refresh_token": req.RefreshToken, "error": err}).Error("Failed to Refresh Access Token!")
		return nil, status.Errorf(codes.Internal, "failed to refresh access token: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Access Token Refreshed Successfully!")
	return &authProto.LoginUserResponse{
		User: &authProto.User{
			Id:           res.User.ID.Hex(),
			Name:         res.User.Name,
			Username:     res.User.Username,
			Email:        res.User.Email,
			Phone:        res.User.Phone,
			PasswordHash: res.User.PasswordHash,
			Incentive:    res.User.Incentive,
			IsActive:     res.User.IsActive,
			IsSuperAdmin: res.User.IsSuperAdmin,
			CreatedAt:    timestamppb.New(res.User.CreatedAt),
			UpdatedAt:    timestamppb.New(res.User.UpdatedAt),
		},
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}, nil
}

// --- Role Management ---

func (s *authGRPCServer) RegisterRole(ctx context.Context, req *authProto.RoleRegisterPayload) (*authProto.Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.RegisterRole")
	defer span.End()

	s.log.WithFields(logrus.Fields{"name": req.Name}).Info("Registering Role...")

	res, err := s.authService.RegisterRole(ctx, req.Name, req.Description)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"name": req.Name, "error": err}).Error("Failed to Register Role!")
		return nil, status.Errorf(codes.Internal, "failed to register role: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role Registered Successfully!")
	return &authProto.Role{
		Id:        res.ID.Hex(),
		Name:      res.Name,
		CreatedAt: timestamppb.New(res.CreatedAt),
		UpdatedAt: timestamppb.New(res.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) GetRoleByID(ctx context.Context, req *authProto.GetRoleByIdPayload) (*authProto.Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetRoleById")
	defer span.End()

	role, err := s.authService.GetRoleByID(ctx, req.Id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"role_id": req.Id, "error": err}).Error("Failed to Get Role!")
		return nil, status.Errorf(codes.Internal, "failed to get role: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role Found Successfully!")
	return &authProto.Role{
		Id:        role.ID.Hex(),
		Name:      role.Name,
		CreatedAt: timestamppb.New(role.CreatedAt),
		UpdatedAt: timestamppb.New(role.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) GetRoleByIdentifier(ctx context.Context, req *authProto.GetRoleByIdentifierPayload) (*authProto.Role, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetRoleByIdentifier")
	defer span.End()

	role, err := s.authService.GetRoleByIdentifier(ctx, req.Identifier)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"identifier": req.Identifier, "error": err}).Error("Failed to Get Role!")
		return nil, status.Errorf(codes.Internal, "failed to get role: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role Found Successfully!")
	return &authProto.Role{
		Id:        role.ID.Hex(),
		Name:      role.Name,
		CreatedAt: timestamppb.New(role.CreatedAt),
		UpdatedAt: timestamppb.New(role.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) GetAllRole(ctx context.Context, req *emptypb.Empty) (*authProto.RoleList, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetAllRole")
	defer span.End()

	roles, err := s.authService.GetAllRole(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())

		s.log.WithFields(logrus.Fields{"error": err}).Error("Failed to Get All Roles!")
		return nil, status.Errorf(codes.Internal, "failed to get all roles: %v", err)
	}

	rolesProto := make([]*authProto.Role, len(roles))
	for i, role := range roles {
		rolesProto[i] = &authProto.Role{
			Id:        role.ID.Hex(),
			Name:      role.Name,
			CreatedAt: timestamppb.New(role.CreatedAt),
			UpdatedAt: timestamppb.New(role.UpdatedAt),
		}
	}

	span.SetStatus(otlpcodes.Ok, "All Roles Found Successfully!")
	return &authProto.RoleList{Roles: rolesProto}, nil
}

// GetRoles implements the GetRoles RPC method
func (s *authGRPCServer) GetRoles(ctx context.Context, req *emptypb.Empty) (*authProto.RoleList, error) {
	return s.GetAllRole(ctx, req)
}

func (s *authGRPCServer) UpdateRole(ctx context.Context, req *authProto.UpdateRolePayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.UpdateRole")
	defer span.End()

	s.log.WithFields(logrus.Fields{"role_id": req.Id, "name": req.Name}).Info("Updating Role...")

	_, err := s.authService.UpdateRole(ctx, req.Id, req.Name, req.Description)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		s.log.WithFields(logrus.Fields{"role_id": req.Id, "error": err}).Error("Failed to Update Role!")
		return nil, status.Errorf(codes.Internal, "failed to update role: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role Updated Successfully!")
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) DeleteRole(ctx context.Context, req *authProto.DeleteRolePayload) (*emptypb.Empty, error) {
	// Not implemented in service yet
	return &emptypb.Empty{}, nil
}

// --- Enhanced RBAC Handlers ---

func (s *authGRPCServer) AssignPermissionsToRole(ctx context.Context, req *authProto.AssignPermissionsToRolePayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.AssignPermissionsToRole")
	defer span.End()

	s.log.WithFields(logrus.Fields{"role_id": req.RoleId, "permissions_count": len(req.PermissionIds)}).Info("Assigning Permissions to Role...")

	err := s.authService.AssignPermissionsToRole(ctx, req.RoleId, req.PermissionIds)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		s.log.WithFields(logrus.Fields{"role_id": req.RoleId, "error": err}).Error("Failed to Assign Permissions to Role!")
		return nil, status.Errorf(codes.Internal, "failed to assign permissions: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Permissions Assigned Successfully!")
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) AssignRoleToUser(ctx context.Context, req *authProto.AssignRoleToUserPayload) (*authProto.User, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.AssignRoleToUser")
	defer span.End()

	s.log.WithFields(logrus.Fields{"user_id": req.UserId, "role_id": req.RoleId}).Info("Assigning Role to User...")

	user, err := s.authService.AssignRoleToUser(ctx, req.UserId, req.RoleId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		s.log.WithFields(logrus.Fields{"user_id": req.UserId, "role_id": req.RoleId, "error": err}).Error("Failed to Assign Role to User!")
		return nil, status.Errorf(codes.Internal, "failed to assign role to user: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role Assigned Successfully!")
	return &authProto.User{
		Id:           user.ID.Hex(),
		Name:         user.Name,
		Username:     user.Username,
		Email:        user.Email,
		Phone:        user.Phone,
		IsActive:     user.IsActive,
		IsSuperAdmin: user.IsSuperAdmin,
		CompanyIds:   user.CompanyIDs,
		CreatedAt:    timestamppb.New(user.CreatedAt),
		UpdatedAt:    timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *authGRPCServer) RemovePermissionFromRole(ctx context.Context, req *authProto.RemovePermissionFromRolePayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.RemovePermissionFromRole")
	defer span.End()

	s.log.WithFields(logrus.Fields{"role_id": req.RoleId, "permission_id": req.PermissionId}).Info("Removing Permission from Role...")

	err := s.authService.RemovePermissionFromRole(ctx, req.RoleId, req.PermissionId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		s.log.WithFields(logrus.Fields{"role_id": req.RoleId, "permission_id": req.PermissionId, "error": err}).Error("Failed to Remove Permission from Role!")
		return nil, status.Errorf(codes.Internal, "failed to remove permission: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Permission Removed Successfully!")
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) RemoveRoleFromUser(ctx context.Context, req *authProto.RemoveRoleFromUserPayload) (*emptypb.Empty, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.RemoveRoleFromUser")
	defer span.End()

	s.log.WithFields(logrus.Fields{"user_id": req.UserId, "role_id": req.RoleId}).Info("Removing Role from User...")

	err := s.authService.RemoveRoleFromUser(ctx, req.UserId, req.RoleId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		s.log.WithFields(logrus.Fields{"user_id": req.UserId, "role_id": req.RoleId, "error": err}).Error("Failed to Remove Role from User!")
		return nil, status.Errorf(codes.Internal, "failed to remove role from user: %v", err)
	}

	span.SetStatus(otlpcodes.Ok, "Role Removed Successfully!")
	return &emptypb.Empty{}, nil
}

func (s *authGRPCServer) GetRolePermissions(ctx context.Context, req *authProto.GetRolePermissionsPayload) (*authProto.GetRolePermissionsResponse, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetRolePermissions")
	defer span.End()

	s.log.WithFields(logrus.Fields{"role_id": req.RoleId}).Info("Getting Role Permissions...")

	permissions, err := s.authService.GetRolePermissions(ctx, req.RoleId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		s.log.WithFields(logrus.Fields{"role_id": req.RoleId, "error": err}).Error("Failed to Get Role Permissions!")
		return nil, status.Errorf(codes.Internal, "failed to get role permissions: %v", err)
	}

	var pbPermissions []*authProto.Permission
	for _, p := range permissions {
		pbPermissions = append(pbPermissions, &authProto.Permission{
			Name:        p.Name,
			Description: p.Description,
		})
	}

	span.SetStatus(otlpcodes.Ok, "Role Permissions Retrieved Successfully!")
	return &authProto.GetRolePermissionsResponse{Permissions: pbPermissions}, nil
}

func (s *authGRPCServer) GetUserPermissions(ctx context.Context, req *authProto.GetUserPermissionsPayload) (*authProto.GetUserPermissionsResponse, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.GetUserPermissions")
	defer span.End()

	s.log.WithFields(logrus.Fields{"user_id": req.UserId}).Info("Getting User Permissions...")

	permissions, err := s.authService.GetUserPermissions(ctx, req.UserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otlpcodes.Error, err.Error())
		s.log.WithFields(logrus.Fields{"user_id": req.UserId, "error": err}).Error("Failed to Get User Permissions!")
		return nil, status.Errorf(codes.Internal, "failed to get user permissions: %v", err)
	}

	var pbPermissions []*authProto.Permission
	for _, p := range permissions {
		pbPermissions = append(pbPermissions, &authProto.Permission{
			Name:        p.Name,
			Description: p.Description,
		})
	}

	span.SetStatus(otlpcodes.Ok, "User Permissions Retrieved Successfully!")
	return &authProto.GetUserPermissionsResponse{Permissions: pbPermissions}, nil
}

func (s *authGRPCServer) CheckUserPermission(ctx context.Context, req *authProto.CheckPermissionRequest) (*authProto.CheckPermissionResponse, error) {
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "authGRPCServer.CheckUserPermission")
	defer span.End()

	allowed, err := s.authService.CheckUserPermission(ctx, req.UserId, req.Resource, req.Action)
	if err != nil {
		// Log error but generally fail safe
		span.RecordError(err)
		s.log.WithFields(logrus.Fields{"user_id": req.UserId, "error": err}).Error("Failed to Check Permission!")
		return &authProto.CheckPermissionResponse{Allowed: false, Message: err.Error()}, nil
	}

	span.SetStatus(otlpcodes.Ok, "Permission Checked Successfully!")
	return &authProto.CheckPermissionResponse{Allowed: allowed}, nil
}
