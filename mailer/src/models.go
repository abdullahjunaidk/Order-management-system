package src

// UserRegisteredEventPayload struct.
// This struct is used to define the user registered event payload.
//
// Attributes:
//   - ActivationToken (string): The activation token.
//   - Username (string): The username.
//   - Email (string): The email.
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
