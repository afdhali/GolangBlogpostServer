package dto

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,username"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	FullName string `json:"full_name" validate:"omitempty,max=100"`
	Avatar   string `json:"avatar" validate:"omitempty,url"`
	Role     string `json:"role" validate:"required,oneof=super_admin admin user"`
}

type UpdateUserRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50,username"`
	Email    string `json:"email" validate:"omitempty,email"`
	FullName string `json:"full_name" validate:"omitempty,max=100"`
	Password string `json:"password" validate:"omitempty,min=8,max=100"`
	Avatar   string `json:"avatar" validate:"omitempty,url"`
	Role     string `json:"role" validate:"omitempty,oneof=super_admin admin user"`
	IsActive *bool  `json:"is_active" validate:"omitempty"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" validate:"omitempty,min=3,max=50,username"`
	FullName string `json:"full_name" validate:"omitempty,max=100"`
	Avatar   string `json:"avatar" validate:"omitempty,url"`
}

type UserQueryParams struct {
	Page      int    `form:"page" validate:"omitempty,min=1"`
	Limit     int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Search    string `form:"search" validate:"omitempty,max=100"`
	Role      string `form:"role" validate:"omitempty,oneof=super_admin admin user"`
	IsActive  *bool  `form:"is_active" validate:"omitempty"`
	SortBy    string `form:"sort_by" validate:"omitempty,oneof=created_at updated_at username email"`
	SortOrder string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
}