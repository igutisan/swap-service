package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// JWTSecret se debe configurar desde variables de entorno en producción
	JWTSecret = []byte("tu-clave-secreta-aqui-cambia-en-produccion")
	// TokenExpiry tiempo de expiración del token (24 horas)
	TokenExpiry = 24 * time.Hour
	// ActorTypeUser representa un usuario final en el sistema.
	ActorTypeUser = "user"
	// ActorTypeCompany representa una empresa en el sistema.
	ActorTypeCompany = "company"
)

// AuthClaims representa los claims unificados del JWT para cualquier actor autenticado. 
type AuthClaims struct {
	Email     string `json:"email"`
	ActorType string `json:"actor_type"`

	// Campos legacy para compatibilidad temporal con tokens antiguos.
	UserID    uuid.UUID `json:"user_id,omitempty"`
	CompanyID uuid.UUID `json:"company_id,omitempty"`
	jwt.RegisteredClaims
}

// GenerateAuthJWT genera un nuevo token JWT para un actor autenticado.
func GenerateAuthJWT(actorID uuid.UUID, email, actorType string) (string, error) {
	claims := AuthClaims{
		Email:     email,
		ActorType: actorType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuerByActorType(actorType),
			Subject:   actorID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", fmt.Errorf("error al firmar token: %w", err)
	}

	return tokenString, nil
}

// ValidateAuthJWT valida un token JWT y retorna los claims normalizados.
func ValidateAuthJWT(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expirado")
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, errors.New("token malformado")
		}
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		normalized, normalizeErr := normalizeAuthClaims(claims)
		if normalizeErr != nil {
			return nil, normalizeErr
		}

		return normalized, nil
	}

	return nil, errors.New("claims inválidos")
}

// ExtractUserIDFromToken extrae el userID de un token JWT unificado.
func ExtractUserIDFromToken(tokenString string) (uuid.UUID, error) {
	claims, err := ValidateAuthJWT(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	if claims.ActorType != ActorTypeUser {
		return uuid.Nil, errors.New("token no corresponde a un usuario")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("subject inválido en token")
	}

	return userID, nil
}

// ExtractCompanyIDFromToken extrae el companyID de un token JWT unificado.
func ExtractCompanyIDFromToken(tokenString string) (uuid.UUID, error) {
	claims, err := ValidateAuthJWT(tokenString)
	if err != nil {
		return uuid.Nil, err
	}

	if claims.ActorType != ActorTypeCompany {
		return uuid.Nil, errors.New("token no corresponde a una empresa")
	}

	companyID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("subject inválido en token")
	}

	return companyID, nil
}

func normalizeAuthClaims(claims *AuthClaims) (*AuthClaims, error) {
	if claims == nil {
		return nil, errors.New("claims inválidos")
	}

	if strings.TrimSpace(claims.Subject) == "" {
		switch {
		case claims.UserID != uuid.Nil:
			claims.Subject = claims.UserID.String()
		case claims.CompanyID != uuid.Nil:
			claims.Subject = claims.CompanyID.String()
		default:
			return nil, errors.New("subject inválido en token")
		}
	}

	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, errors.New("subject inválido en token")
	}

	if claims.ActorType == "" {
		inferredType, err := inferActorTypeFromLegacyClaims(claims)
		if err != nil {
			return nil, err
		}
		claims.ActorType = inferredType
	}

	if claims.ActorType != ActorTypeUser && claims.ActorType != ActorTypeCompany {
		return nil, errors.New("actor_type inválido en token")
	}

	return claims, nil
}

func inferActorTypeFromLegacyClaims(claims *AuthClaims) (string, error) {
	issuer := strings.TrimSpace(claims.Issuer)
	if issuer == "swap-service" {
		return ActorTypeUser, nil
	}
	if issuer == "swap-service-company" {
		return ActorTypeCompany, nil
	}

	if claims.UserID != uuid.Nil {
		return ActorTypeUser, nil
	}
	if claims.CompanyID != uuid.Nil {
		return ActorTypeCompany, nil
	}

	return "", errors.New("no se pudo inferir actor_type desde token legado")
}

func jwtIssuerByActorType(actorType string) string {
	if actorType == ActorTypeCompany {
		return "swap-service-company"
	}

	return "swap-service"
}
