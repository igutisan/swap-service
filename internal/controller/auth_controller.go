package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"swap/iguti/swap-service/internal/domain"
	"swap/iguti/swap-service/internal/dto"
	"swap/iguti/swap-service/internal/middleware"
	"swap/iguti/swap-service/internal/utils"
)

// AuthController maneja los endpoints de autenticación
type AuthController struct {
	authService domain.AuthService
}

// NewAuthController crea una nueva instancia de AuthController
func NewAuthController(authService domain.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// @Summary Registra un nuevo usuario
// @Description Crea una nueva cuenta de usuario con email y contraseña
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterUserRequest true "Datos de registro"
// @Success 201 {object} dto.APIResponse{data=dto.RegisterUserData}
// @Failure 400 {object} dto.APIResponse
// @Failure 409 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	// Obtener request validado del contexto (inyectado por el middleware)
	req := middleware.GetValidatedRequest(ctx).(*dto.RegisterUserRequest)

	// Llamar al servicio
	user, err := c.authService.RegisterUser(ctx, req.Name, req.Email, req.Password, req.Phone)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExist):
			resp := dto.ErrorResponse(409, "EMAIL_EXISTS: El email ya está registrado")
			ctx.JSON(http.StatusConflict, resp)
		default:
			resp := dto.ErrorResponse(500, "INTERNAL_ERROR: Error interno del servidor", dto.WithErrors(err.Error()))
			ctx.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	// Construir respuesta exitosa
	data := dto.RegisterUserData{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Usuario registrado exitosamente"),
	)
	ctx.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Inicia sesión de usuario
// @Description Autentica un usuario y retorna un token JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Credenciales de login"
// @Success 200 {object} dto.APIResponse{data=dto.LoginData}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	// Obtener request validado del contexto
	req := middleware.GetValidatedRequest(ctx).(*dto.LoginRequest)

	// Llamar al servicio
	user, token, err := c.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrInvalidPassword):
			resp := dto.ErrorResponse(401, "INVALID_CREDENTIALS: Email o contraseña incorrectos")
			ctx.JSON(http.StatusUnauthorized, resp)
		default:
			resp := dto.ErrorResponse(500, "INTERNAL_ERROR: Error interno del servidor", dto.WithErrors(err.Error()))
			ctx.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	// Construir respuesta exitosa
	data := dto.LoginData{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: 86400,
		ActorType: utils.ActorTypeUser,
		User: dto.UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
		},
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Inicio de sesión exitoso"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// SetupRoutes configura las rutas de autenticación con middleware de validación
func (c *AuthController) SetupRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		// Rutas de usuario
		auth.POST("/register", middleware.ValidationMiddleware(&dto.RegisterUserRequest{}), c.Register)
		auth.POST("/login", middleware.ValidationMiddleware(&dto.LoginRequest{}), c.Login)

		// Rutas de empresa
		auth.POST("/company/register", middleware.ValidationMiddleware(&dto.RegisterCompanyRequest{}), c.RegisterCompany)
		auth.POST("/company/login", middleware.ValidationMiddleware(&dto.LoginCompanyRequest{}), c.LoginCompany)
	}
}

// SetupProfileRoutes configura las rutas de perfil con middlewares de autenticación
func (c *AuthController) SetupProfileRoutes(router *gin.RouterGroup, userAuthMiddleware, companyAuthMiddleware gin.HandlerFunc) {
	// Rutas de usuario (protegidas con AuthMiddleware)
	user := router.Group("/user")
	user.Use(userAuthMiddleware)
	{
		user.GET("/profile", c.GetUserProfile)
	}

	// Rutas de empresa (protegidas con CompanyAuthMiddleware)
	company := router.Group("/company")
	company.Use(companyAuthMiddleware)
	{
		company.GET("/profile", c.GetCompanyProfile)
	}
}

// RegisterCompany godoc
// @Summary Registra una nueva empresa
// @Description Crea una nueva cuenta de empresa con email y contraseña
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterCompanyRequest true "Datos de registro de empresa"
// @Success 201 {object} dto.APIResponse{data=dto.RegisterCompanyData}
// @Failure 400 {object} dto.APIResponse
// @Failure 409 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /auth/company/register [post]
func (c *AuthController) RegisterCompany(ctx *gin.Context) {
	req := middleware.GetValidatedRequest(ctx).(*dto.RegisterCompanyRequest)

	company, err := c.authService.RegisterCompany(ctx, req.Name, req.Email, req.Password, req.Phone, req.Address)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExist):
			resp := dto.ErrorResponse(409, "EMAIL_EXISTS: El email ya está registrado")
			ctx.JSON(http.StatusConflict, resp)
		case errors.Is(err, domain.ErrGeocodingNotFound):
			resp := dto.ErrorResponse(400, "ADDRESS_NOT_FOUND: No se pudo encontrar la dirección. Por favor verifica e intenta con una dirección más específica.")
			ctx.JSON(http.StatusBadRequest, resp)
		case errors.Is(err, domain.ErrGeocodingEmptyAddress):
			resp := dto.ErrorResponse(400, "EMPTY_ADDRESS: La dirección no puede estar vacía")
			ctx.JSON(http.StatusBadRequest, resp)
		default:
			resp := dto.ErrorResponse(500, "INTERNAL_ERROR: Error interno del servidor", dto.WithErrors(err.Error()))
			ctx.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	data := dto.RegisterCompanyData{
		ID:        company.ID,
		Name:      company.Name,
		Email:     company.Email,
		Phone:     company.Phone,
		Address:   company.Address,
		Lat:       company.Lat,
		Lng:       company.Lng,
		CreatedAt: company.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Empresa registrada exitosamente"),
	)
	ctx.JSON(http.StatusCreated, resp)
}

// LoginCompany godoc
// @Summary Inicia sesión de empresa
// @Description Autentica una empresa y retorna un token JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginCompanyRequest true "Credenciales de login de empresa"
// @Success 200 {object} dto.APIResponse{data=dto.CompanyLoginData}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /auth/company/login [post]
func (c *AuthController) LoginCompany(ctx *gin.Context) {
	req := middleware.GetValidatedRequest(ctx).(*dto.LoginCompanyRequest)

	company, token, err := c.authService.LoginCompany(ctx, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCompanyNotFound), errors.Is(err, domain.ErrInvalidPassword):
			resp := dto.ErrorResponse(401, "INVALID_CREDENTIALS: Email o contraseña incorrectos")
			ctx.JSON(http.StatusUnauthorized, resp)
		default:
			resp := dto.ErrorResponse(500, "INTERNAL_ERROR: Error interno del servidor", dto.WithErrors(err.Error()))
			ctx.JSON(http.StatusInternalServerError, resp)
		}
		return
	}

	data := dto.CompanyLoginData{
		Token:     token,
		TokenType: "Bearer",
		ExpiresIn: 86400,
		ActorType: utils.ActorTypeCompany,
		Company: dto.CompanyInfo{
			ID:      company.ID,
			Name:    company.Name,
			Email:   company.Email,
			Phone:   company.Phone,
			Address: company.Address,
			Lat:     company.Lat,
			Lng:     company.Lng,
			LogoURL: company.LogoURL,
		},
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Inicio de sesión exitoso"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// GetUserProfile godoc
// @Summary Obtiene el perfil del usuario autenticado
// @Description Retorna la información del perfil del usuario actual
// @Tags user
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.UserProfileData}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /user/profile [get]
func (c *AuthController) GetUserProfile(ctx *gin.Context) {
	user, exists := ctx.Get("user")
	if !exists {
		resp := dto.ErrorResponse(401, "UNAUTHORIZED: Usuario no autenticado")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	userData := user.(*domain.User)

	data := dto.UserProfileData{
		ID:        userData.ID,
		Name:      userData.Name,
		Email:     userData.Email,
		Phone:     userData.Phone,
		CreatedAt: userData.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Perfil obtenido exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// GetCompanyProfile godoc
// @Summary Obtiene el perfil de la empresa autenticada
// @Description Retorna la información del perfil de la empresa actual
// @Tags company
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.CompanyProfileData}
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /company/profile [get]
func (c *AuthController) GetCompanyProfile(ctx *gin.Context) {
	company, exists := ctx.Get("company")
	if !exists {
		resp := dto.ErrorResponse(401, "UNAUTHORIZED: Empresa no autenticada")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	companyData := company.(*domain.Company)

	data := dto.CompanyProfileData{
		ID:        companyData.ID,
		Name:      companyData.Name,
		Email:     companyData.Email,
		Phone:     companyData.Phone,
		Address:   companyData.Address,
		Lat:       companyData.Lat,
		Lng:       companyData.Lng,
		LogoURL:   companyData.LogoURL,
		CreatedAt: companyData.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Perfil obtenido exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}
