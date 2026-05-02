package dto

type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100" example:"Alice Smith"`
	Email    string `json:"email"     validate:"required,email"          example:"alice@example.com"`
	Password string `json:"password"  validate:"required,min=8"          example:"secret1234"`
	Role     string `json:"role"      validate:"required,oneof=candidate recruiter" example:"candidate"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email" example:"alice@example.com"`
	Password string `json:"password" validate:"required"       example:"secret1234"`
}

type AuthResponse struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string `json:"id"                   example:"d4f5a2b1-..."`
	FullName  string `json:"full_name"            example:"Alice Smith"`
	Email     string `json:"email"                example:"alice@example.com"`
	Role      string `json:"role"                 example:"candidate"`
	CompanyID string `json:"company_id,omitempty" example:""`
	CreatedAt string `json:"created_at"           example:"2026-05-01T12:00:00Z"`
}
