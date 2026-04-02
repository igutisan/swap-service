package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	// JWTSecret se debe configurar desde variables de entorno en producción
	JWTSecret = []byte("tu-clave-secreta-aqui-cambia-en-produccion")
	// TokenExpiry tiempo de expiración del token (24 horas)
	TokenExpiry = 24 * time.Hour
)

// JWTClaims representa los claims personalizados del JWT para usuarios
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// CompanyJWTClaims representa los claims personalizados del JWT para empresas
type CompanyJWTClaims struct {
	CompanyID uuid.UUID `json:"company_id"`
	Email     string    `json:"email"`
	jwt.RegisteredClaims
}

// GenerateJWT genera un nuevo token JWT para un usuario
func GenerateJWT(userID uuid.UUID, email string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "swap-service",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", fmt.Errorf("error al firmar token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT valida un token JWT y retorna los claims
func ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
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

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("claims inválidos")
}

// ExtractUserIDFromToken extrae el userID de un token JWT
func ExtractUserIDFromToken(tokenString string) (uuid.UUID, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// GenerateCompanyJWT genera un nuevo token JWT para una empresa
func GenerateCompanyJWT(companyID uuid.UUID, email string) (string, error) {
	claims := CompanyJWTClaims{
		CompanyID: companyID,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "swap-service-company",
			Subject:   companyID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", fmt.Errorf("error al firmar token: %w", err)
	}

	return tokenString, nil
}

// ValidateCompanyJWT valida un token JWT de empresa y retorna los claims
func ValidateCompanyJWT(tokenString string) (*CompanyJWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CompanyJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
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

	if claims, ok := token.Claims.(*CompanyJWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("claims inválidos")
}

// ExtractCompanyIDFromToken extrae el companyID de un token JWT de empresa
func ExtractCompanyIDFromToken(tokenString string) (uuid.UUID, error) {
	claims, err := ValidateCompanyJWT(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.CompanyID, nil
}
