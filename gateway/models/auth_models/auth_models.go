package auth_models

import (
	"time"
)

// Customer struct.
// This struct is used to represent a customer.
//
// Attributes:
//   - ID (string): The ID.
//   - Name (string): The name.
//   - Phone (int64): The phone.
//   - PostOffice (string): The post office.
//   - PinCode (int32): The pin code.
//   - State (string): The state.
//   - District (string): The district.
type Customer struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name" validate:"required" example:"johndoe"`
	Phone      int64  `json:"phone" validate:"required,min=6000000000,max=9999999999" example:"9876543210"`
	PostOffice string
	PinCode    int32
	State      string
	District   string
}

// ErrorResponse struct.
// This struct is used to represent an error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// ----- Company ----- //

// Company struct.
// This struct is used to represent a company.
//
// Attributes:
//   - ID (string): The ID.
//   - Name (string): The name.
//   - Email (string): The email.
//   - CreatedAt (time.Time): The created at.
//   - UpdatedAt (time.Time): The updated at.
type Company struct {
	ID          string    `json:"id"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CompanyRegisterPayload struct {
	Name        string `json:"name" validate:"required,min=3,max=20" example:"Nike"`
	Description string `json:"description,omitempty"`
}

type CompanyRegisterSuccessResponse struct {
	Message string  `json:"message" example:"Company Registered Successfully!"`
	Company Company `json:"company"`
}

type CompanyRegisterErrorResponse struct {
	Message string `json:"message" example:"Failed to Register Company!"`
	Error   string `json:"error" example:"<error_message>"`
}

// ResendActivationEmailPayload struct.
// This struct is used to represent a user resend activation email payload.
//
// Attributes:
//   - Identifier (string): The identifier.
type ResendActivationEmailPayload struct {
	Identifier string `json:"identifier" validate:"required" example:"johndoe/johndoe@example.com"`
}

// ResendActivationEmailSuccessResponse struct.
// This struct is used to represent a resend activation email success response.
//
// Attributes:
//   - Message (string): The message.
//   - User (User): The user.
type ResendActivationEmailSuccessResponse struct {
	Message string `json:"message" example:"Activation Email Sent Successfully!"`
	User    User   `json:"user"`
}

// ResendActivationEmailErrorResponse struct.
// This struct is used to represent a resend activation email error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type ResendActivationEmailErrorResponse struct {
	Message string `json:"message" example:"Failed to Send Activation Email!"`
	Error   string `json:"error" example:"<error_message>"`
}

// ForgotPasswordPayload struct.
// This struct is used to represent a forgot password payload.
//
// Attributes:
//   - Identifier (string): The identifier.
type ForgotPasswordPayload struct {
	Identifier string `json:"identifier" validate:"required" example:"johndoe/johndoe@example.com"`
}

// ForgotPasswordSuccessResponse struct.
// This struct is used to represent a forgot password success response.
//
// Attributes:
//   - Message (string): The message.
type ForgotPasswordSuccessResponse struct {
	Message string `json:"message" example:"Password Reset Email Sent Successfully!"`
}

// ForgotPasswordErrorResponse struct.
// This struct is used to represent a forgot password error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type ForgotPasswordErrorResponse struct {
	Message string `json:"message" example:"Failed to Send Password Reset Email!"`
	Error   string `json:"error" example:"<error_message>"`
}

// ResetPasswordPayload struct.
// This struct is used to represent a reset password payload.
//
// Attributes:
//   - Password (string): The password.
//   - ConfrimPassword (string): The confirm password.
type ResetPasswordPayload struct {
	Password        string `json:"password" validate:"required,min=8" example:"strong@password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password" example:"strong@password"`
}

// ResetPasswordSuccessResponse struct.
// This struct is used to represent a reset password success response.
//
// Attributes:
//   - Message (string): The message.
type ResetPasswordSuccessResponse struct {
	Message string `json:"message" example:"Password Reset Successfully!"`
}

// ResetPasswordErrorResponse struct.
// This struct is used to represent a reset password error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type ResetPasswordErrorResponse struct {
	Message string `json:"message" example:"Failed to Reset Password!"`
	Error   string `json:"error" example:"<error_message>"`
}

// UnauthorizedErrorResponse struct.
// This struct is used to represent an unauthorized error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type UnauthorizedErrorResponse struct {
	Message string `json:"message" example:"Unauthorized!"`
	Error   string `json:"error" example:"<error_message>"`
}

type BaseModel struct {
	ID        string    `json:"id,omitempty" example:"67adca0b4e864cfffb002299"`
	CreatedAt time.Time `json:"createdAt" example:"2021-07-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updatedAt,omitempty" example:"2021-07-01T00:00:00Z"`
}
