package src

import (
	"common/types"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Core Entities ---

type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name               string             `bson:"name" json:"name"`
	Username           string             `bson:"username" json:"username"`
	Email              string             `bson:"email" json:"email"`
	PasswordHash       string             `bson:"password_hash" json:"-"`
	Phone              int64              `bson:"phone" json:"phone"`
	Incentive          int64              `bson:"incentive" json:"incentive"`
	IsActive           bool               `bson:"isActive" json:"isActive"`
	IsSuperAdmin       bool               `bson:"isSuperAdmin" json:"isSuperAdmin" default:"false"`
	CompanyIDs         []string           `bson:"companyId" json:"companyId"`
	PasswordResetToken string             `bson:"passwordResetToken" json:"-"`
	AccessToken        string             `bson:"accessToken" json:"-"`
	RefreshToken       string             `bson:"refreshToken" json:"-"`
	CreatedAt          time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt          time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Role struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Permission struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type Company struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	IsActive    bool               `json:"isActive" bson:"isActive"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type UserRole struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    string             `json:"userId" bson:"userId"`
	RoleID    string             `json:"roleId" bson:"roleId"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type RolePermission struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PermissionID string             `json:"permissionId" bson:"permissionId"`
	RoleID       string             `json:"roleId" bson:"roleId"`
	CreatedAt    time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt" bson:"updatedAt"`
}

// --- User Payloads ---

type RegisterUserPayload struct {
	Name         string   `json:"name"`
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	Phone        int64    `json:"phone"`
	PasswordHash string   `json:"password_hash"`
	CompanyIDs   []string `json:"company_ids"`
}

type GetUserByIDPayload struct {
	UserID string `json:"user_id"`
}

type GetUserByIdentifierPayload struct {
	Identifier string `json:"identifier"`
}

type ActivateUserPayload struct {
	ActivationToken string `json:"activationToken"`
}

type ResendActivationPayload struct {
	Identifier string `json:"identifier"`
}

type UserPasswordResetPayload struct {
	PasswordResetToken string `json:"password_reset_token"`
	PasswordHash       string `json:"password_hash"`
}

type UserVerifyAccessTokenPayload struct {
	AccessToken string `json:"access_token"`
}

type UserRefreshAccessTokenPayload struct {
	RefreshToken string `json:"refresh_token"`
}

type GrantCompanyAccessPayload struct {
	UserID     string   `json:"user_id"`
	CompanyIDs []string `json:"company_ids"`
}

type UserLoginResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// --- Role & Permission Payloads ---

type RoleRegisterPayload struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Permissions []types.Permission `json:"permissions"`
}

type PermissionRegisterPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UserRoleRegisterPayload struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

type PermissionRoleRegisterPayload struct {
	PermissionID string `json:"permission_id"`
	RoleID       string `json:"role_id"`
}

// --- Company Payloads ---

type CompanyRegisterPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CompanySuccessResponse struct {
	Message string
	Company Company
}

// --- Event Payloads ---

type UserRegisteredEventPayload struct {
	ActivationToken string `json:"activation_token"`
	Username        string `json:"username"`
	Email           string `json:"email"`
}

// UserActivatedEventPayload struct.
// This struct is used to define the user activated event payload.
//
// Attributes:
//   - Username (string): The username.
//   - Email (string): The email.
type UserActivatedEventPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// UserForgotPasswordEventPayload struct.
// This struct is used to define the password reset event payload.
//
// Attributes:
//   - PasswordResetToken (string): The password reset token.
//   - Username (string): The username.
//   - Email (string): The email.
type UserForgotPasswordEventPayload struct {
	PasswordResetToken string `json:"password_reset_token"`
	Username           string `json:"username"`
	Email              string `json:"email"`
}

// UserPasswordResetEventPayload struct.
// This struct is used to define the password reset event payload.
//
// Attributes:
//   - Username (string): The username.
//   - Email (string): The email.
type UserPasswordResetEventPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}
