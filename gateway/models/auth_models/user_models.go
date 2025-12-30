package auth_models

import "time"

// Employee struct.
// This struct is used to represent an employee.
//
// Attributes:
//   - ID (string): The ID.
//   - UserName (string): The user name.
//   - Name (string): The name.
//   - Phone (int64): The phone.
//   - Email (string): The email.
//   - Incentive (int64): The incentive.
//   - IsActive (bool): The is active.
type User struct {
	ID                 string       `json:"id"`
	Name               string       `json:"name" validate:"required" example:"johndoe"`
	Username           string       `json:"username" validate:"required,min=3,max=20" example:"johndoe"`
	Email              string       `json:"email" validate:"required,email" example:"johndoe@example.com"`
	Phone              int64        `json:"phone" validate:"required, len=10" example:"9876543210"`
	Incentive          int64        `json:"incentive"`
	IsSuperAdmin       bool         `json:"is_super_admin"`
	IsActive           bool         `json:"is_active"`
	CompanyIds         []string     `json:"company_ids"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
	ActivationToken    string       `json:"-" validate:"omitempty"`
	PasswordResetToken string       `json:"-" validate:"omitempty"`
}

// EmployeeRegisterPayload struct.
// This struct is used to represent an employee register payload.
//
// Attributes:
//   - UserName (string): The user name.
//   - Name (string): The name.
//   - Email (string): The email.
//   - Phone (int64): The phone.
//   - PasswordHash (string): The password hash.
type UserRegisterPayload struct {
	UserName     string   `json:"username" validate:"required,min=3,max=20" example:"johndoeee"`
	Name         string   `json:"name" validate:"required" example:"johndoe"`
	Email        string   `json:"email" validate:"required,email" example:"johndoe@example.com"`
	Phone        int64    `json:"phone" validate:"required,len=10" example:"9876543210"`
	PasswordHash string   `json:"password" validate:"required,min=8" example:"strong@password"`
	CompanyIds   []string `json:"companyIds"`
}

// UserRegisterSuccessResponse struct.
// This struct is used to represent a user register success response.
//
// Attributes:
//   - Message (string): The message.
//   - User (User): The user.
type UserRegisterSuccessResponse struct {
	Message string `json:"message" example:"User Registered Successfully!"`
	User    User   `json:"user"`
}

// UserLoginPayload struct.
// This struct is used to represent a user login payload.
//
// Attributes:
//   - Identifier (string): The identifier.
//   - Password (string): The password.
type UserLoginPayload struct {
	Identifier string `json:"identifier" validate:"required" example:"johndoe/johndoe@example.com"`
	Password   string `json:"password" validate:"required" example:"strong@password"`
}

type UserLoginSuccessResponse struct {
	Message      string `json:"message" example:"User Logged In Successfully!"`
	User         User   `json:"user"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjdhZGNhMGI0ZTg2NGNmZmIwMDIyOTkiLCJleHBpcmVkX2F0IjoiMjAyMS0wNy0wMVQwMDowMDowMC4wMDBaIiwiaWF0IjoxNjI0MzQwNjAwfQ.7"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjdhZGNhMGI0ZTg2NGNmZmIwMDIyOTkiLCJleHBpcmVkX2F0IjoiMjAyMS0wNy0wMVQwMDowMDowMC4wMDBaIiwiaWF0IjoxNjI0MzQwNjAwfQ.7"`
}

// UserLogoutSuccessResponse struct.
// This struct is used to represent a user logout success response.
//
// Attributes:
//   - Message (string): The message.
type UserLogoutSuccessResponse struct {
	Message string `json:"message" example:"User Logged Out Successfully!"`
}

// UserLogoutErrorResponse struct.
// This struct is used to represent a user logout error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type UserErrorResponse struct {
	Message string `json:"message" example:"Failed to Logout User!"`
	Error   string `json:"error" example:"<error_message>"`
}

// UserRegisterErrorResponse struct.
// This struct is used to represent a user register error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type UserRegisterErrorResponse struct {
	Message string `json:"message" example:"Failed to Register User!"`
	Error   string `json:"error" example:"<error_message>"`
}

// GetUserByIDSuccessResponse struct.
// This struct is used to represent a get user by ID success response.
//
// Attributes:
//   - Message (string): The message.
//   - User (User): The user.
type GetUserByIDSuccessResponse struct {
	Message string `json:"message" example:"User Found Successfully!"`
	User    User   `json:"user"`
}

// GetUserByIDErrorResponse struct.
// This struct is used to represent a get user by ID error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type GetUserByIDErrorResponse struct {
	Message string `json:"message" example:"Failed to Find User!"`
	Error   string `json:"error" example:"<error_message>"`
}

// UserActivateSuccessResponse struct.
// This struct is used to represent a user activate success response.
//
// Attributes:
//   - Message (string): The message.
//   - User (User): The user.
type UserActivateSuccessResponse struct {
	Message string `json:"message" example:"User Activated Successfully!"`
	User    User   `json:"user"`
}

// UserActivateErrorResponse struct.
// This struct is used to represent a user activate error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type UserActivateErrorResponse struct {
	Message string `json:"message" example:"Failed to Activate User!"`
	Error   string `json:"error" example:"<error_message>"`
}

type UserLogoutPayload struct {
	Identifier string `json:"identifier" validate:"required" example:"johndoe/johndoe@example.com"`
}

// UserRefreshAccessTokenPayload struct.
// This struct is used to represent an user refresh access token payload.
//
// Attributes:
//   - RefreshToken (string): The refresh token.
type UserRefreshAccessTokenPayload struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjdhZGNhMGI0ZTg2NGNmZmIwMDIyOTkiLCJleHBpcmVkX2F0IjoiMjAyMS0wNy0wMVQwMDowMDowMC4wMDBaIiwiaWF0IjoxNjI0MzQwNjAwfQ.7"`
}

// UserRefreshAccessTokenSuccessResponse struct.
// This struct is used to represent an user refresh access token success response.
//
// Attributes:
//   - Message (string): The message.
//   - User (User): The user.
//   - AccessToken (string): The access token.
//   - RefreshToken (string): The refresh token.
type UserRefreshAccessTokenSuccessResponse struct {
	Message      string `json:"message" example:"User Access Token Refreshed Successfully!"`
	User         User   `json:"user"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjdhZGNhMGI0ZTg2NGNmZmIwMDIyOTkiLCJleHBpcmVkX2F0IjoiMjAyMS0wNy0wMVQwMDowMDowMC4wMDBaIiwiaWF0IjoxNjI0MzQwNjAwfQ.7"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNjdhZGNhMGI0ZTg2NGNmZmIwMDIyOTkiLCJleHBpcmVkX2F0IjoiMjAyMS0wNy0wMVQwMDowMDowMC4wMDBaIiwiaWF0IjoxNjI0MzQwNjAwfQ.7"`
}

// UserRefreshAccessTokenErrorResponse struct.
// This struct is used to represent an user refresh access token error response.
//
// Attributes:
//   - Message (string): The message.
//   - Error (string): The error.
type UserRefreshAccessTokenErrorResponse struct {
	Message string `json:"message" example:"Failed to Refresh Employee Access Token!"`
	Error   string `json:"error" example:"<error_message>"`
}
