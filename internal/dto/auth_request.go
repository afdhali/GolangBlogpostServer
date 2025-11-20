package dto

type RegisterRequest struct {
	Username string `json:"username" validate:"required,username,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	FullName string `json:"full_name" validate:"omitempty,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" validate:"required,min=8"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}