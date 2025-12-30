package auth_models

import "time"

// Admin struct.
// This struct is used to represent an admin.
//
// Attributes:
//   - Username (string): The username.
//   - PasswordResetToken (string): The password reset token.
//   - CreatedAt (time.Time): The created at.
//   - UpdatedAt (time.Time): The updated at.
type Admin struct {
	Username           string    `json:"username" validate:"required,min=3,max=20" example:"johndoe"`
	PasswordResetToken string    `json:"password_reset_token" validate:"omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
