package dto

import "github.com/google/uuid"

// ==================== REQUESTS ====================

// RegisterUserRequest representa la petición para registrar un nuevo usuario
type RegisterUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	Phone    string `json:"phone" validate:"omitempty,max=50"`
}

// LoginRequest representa la petición para iniciar sesión
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

// ==================== DATA RESPONSES ====================

// RegisterUserData representa los datos de respuesta después de registrar un usuario
type RegisterUserData struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt string    `json:"created_at"`
}

// LoginData representa los datos de respuesta después de iniciar sesión
type LoginData struct {
	Token     string   `json:"token"`
	TokenType string   `json:"token_type"`
	ExpiresIn int64    `json:"expires_in"`
	ActorType string   `json:"actor_type"`
	User      UserInfo `json:"user"`
}

// UserInfo representa la información del usuario en la respuesta
type UserInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Phone string    `json:"phone,omitempty"`
}

// ==================== COMPANY REQUESTS ====================

// RegisterCompanyRequest representa la petición para registrar una nueva empresa
type RegisterCompanyRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	Phone    string `json:"phone" validate:"omitempty,max=50"`
	Address  string `json:"address" validate:"required,max=500"`
}

// LoginCompanyRequest representa la petición para iniciar sesión como empresa
type LoginCompanyRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

// ==================== COMPANY DATA RESPONSES ====================

// RegisterCompanyData representa los datos de respuesta después de registrar una empresa
type RegisterCompanyData struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	Address   string    `json:"address"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	CreatedAt string    `json:"created_at"`
}

// CompanyLoginData representa los datos de respuesta después de iniciar sesión como empresa
type CompanyLoginData struct {
	Token     string      `json:"token"`
	TokenType string      `json:"token_type"`
	ExpiresIn int64       `json:"expires_in"`
	ActorType string      `json:"actor_type"`
	Company   CompanyInfo `json:"company"`
}

// CompanyInfo representa la información de la empresa en la respuesta
type CompanyInfo struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Phone   string    `json:"phone,omitempty"`
	Address string    `json:"address"`
	Lat     float64   `json:"lat"`
	Lng     float64   `json:"lng"`
	LogoURL string    `json:"logo_url,omitempty"`
}

// UserProfileData representa los datos del perfil de usuario
type UserProfileData struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt string    `json:"created_at"`
}

// CompanyProfileData representa los datos del perfil de empresa
type CompanyProfileData struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	Address   string    `json:"address"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	LogoURL   string    `json:"logo_url,omitempty"`
	CreatedAt string    `json:"created_at"`
}
