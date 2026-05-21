package controlhub

// CompanyAdminLoginRequest defines the login request payload.
type CompanyAdminLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CompanyAdminLoginResponse defines the login response payload.
type CompanyAdminLoginResponse struct {
	Token    string `json:"token"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	UserType string `json:"user_type"` // Always "controlhub_admin"
}
