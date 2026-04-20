package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"swap/iguti/swap-service/internal/domain"
	"swap/iguti/swap-service/internal/dto"
	"swap/iguti/swap-service/internal/utils"
)

// AuthMiddleware verifica el token JWT y valida que el usuario exista en la base de datos
func AuthMiddleware(userRepo domain.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Token de autorización no proporcionado")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// Verificar que el header comience con "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Formato de autorización inválido. Use: Bearer <token>")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validar el token
		claims, err := utils.ValidateAuthJWT(tokenString)
		if err != nil {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Token inválido", dto.WithErrors(err.Error()))
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		if claims.ActorType != utils.ActorTypeUser {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Tipo de token inválido para esta ruta")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Subject inválido en token")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// VALIDACIÓN CRÍTICA: Verificar que el usuario exista en la base de datos
		user, err := userRepo.GetByID(c.Request.Context(), userID)
		if err != nil {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Usuario no encontrado o no autorizado")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// Verificar que el email del token coincida con el de la base de datos
		if user.Email != claims.Email {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Token inválido - inconsistencia de datos")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// Guardar información del usuario en el contexto
		c.Set("userID", userID.String())
		c.Set("userEmail", claims.Email)
		c.Set("actorType", claims.ActorType)
		c.Set("user", user) // Guardar el objeto usuario completo para uso posterior

		c.Next()
	}
}

// OptionalAuthMiddleware permite peticiones sin token pero extrae info si existe
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateAuthJWT(tokenString)
		if err == nil && claims.ActorType == utils.ActorTypeUser {
			userID, parseErr := uuid.Parse(claims.Subject)
			if parseErr == nil {
				c.Set("userID", userID)
				c.Set("userEmail", claims.Email)
				c.Set("actorType", claims.ActorType)
			}
		}

		if err == nil && claims.ActorType == utils.ActorTypeCompany {
			companyID, parseErr := uuid.Parse(claims.Subject)
			if parseErr == nil {
				c.Set("companyID", companyID.String())
				c.Set("companyEmail", claims.Email)
				c.Set("actorType", claims.ActorType)
			}
		}

		c.Next()
	}
}

// CompanyAuthMiddleware verifica el token JWT de empresa y valida que exista en la base de datos
func CompanyAuthMiddleware(companyRepo domain.CompanyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Obtener el header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Token de autorización no proporcionado")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// Verificar que el header comience con "Bearer "
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Formato de autorización inválido. Use: Bearer <token>")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validar el token de empresa
		claims, err := utils.ValidateAuthJWT(tokenString)
		if err != nil {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Token de empresa inválido", dto.WithErrors(err.Error()))
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		if claims.ActorType != utils.ActorTypeCompany {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Tipo de token inválido para esta ruta")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		companyID, err := uuid.Parse(claims.Subject)
		if err != nil {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Subject inválido en token")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// VALIDACIÓN CRÍTICA: Verificar que la empresa exista en la base de datos
		company, err := companyRepo.GetByID(c.Request.Context(), companyID)
		if err != nil {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Empresa no encontrada o no autorizada")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// Verificar que el email del token coincida con el de la base de datos
		if company.Email != claims.Email {
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: Token inválido - inconsistencia de datos")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		// Guardar información de la empresa en el contexto
		c.Set("companyID", companyID.String())
		c.Set("companyEmail", claims.Email)
		c.Set("actorType", claims.ActorType)
		c.Set("company", company) // Guardar el objeto empresa completo para uso posterior

		c.Next()
	}
}
