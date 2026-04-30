package dto

// RegisterRequest is the expected body for POST /auth/register
type RegisterRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8"`
	Role     string `json:"role"      validate:"required,oneof=candidate recruiter"`
}

// LoginRequest is the expected body for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is returned after successful login
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse is the safe public shape of a user — no password
type UserResponse struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CompanyID string `json:"company_id,omitempty"` // omitempty — omit if empty
	CreatedAt string `json:"created_at"`
}
