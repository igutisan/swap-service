package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Errores de dominio comunes
var (
	ErrUserNotFound              = errors.New("usuario no encontrado")
	ErrCompanyNotFound           = errors.New("compañía no encontrada")
	ErrDishNotFound              = errors.New("plato no encontrado")
	ErrSessionNotFound           = errors.New("sesión no encontrada")
	ErrEmailAlreadyExist         = errors.New("el email ya está registrado")
	ErrInvalidPassword           = errors.New("contraseña inválida")
	ErrValidation                = errors.New("error de validación")
	ErrUnauthorizedSession       = errors.New("no tienes permiso para acceder a esta sesión")
	ErrSessionNotActive          = errors.New("la sesión no está activa")
	ErrInvalidRadius             = errors.New("radio de búsqueda inválido")
	ErrMaxLikesReached           = errors.New("has alcanzado el máximo de 3 likes por sesión")
	ErrDishAlreadySwiped         = errors.New("ya has hecho swipe en este plato")
	ErrDishNotInFinalists        = errors.New("el plato no está entre los finalistas de esta sesión")
	ErrSessionAlreadyCompleted   = errors.New("la sesión ya está completada")
	ErrGeocodingEmptyAddress     = errors.New("la dirección no puede estar vacía")
	ErrGeocodingNotFound         = errors.New("no se encontraron coordenadas para la dirección proporcionada")
	ErrGeocodingInvalidLatitude  = errors.New("la latitud está fuera del rango válido (-90 a 90)")
	ErrGeocodingInvalidLongitude = errors.New("la longitud está fuera del rango válido (-180 a 180)")
)

// TransactionManager define las operaciones para manejar transacciones
type TransactionManager interface {
	BeginTx(ctx context.Context) (interface{}, error)
	Commit(tx interface{}) error
	Rollback(tx interface{}) error
}

// UserRepository define las operaciones de persistencia para usuarios
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// CompanyRepository define las operaciones de persistencia para compañías
type CompanyRepository interface {
	Create(ctx context.Context, company *Company) error
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	GetByEmail(ctx context.Context, email string) (*Company, error)
	Update(ctx context.Context, company *Company) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// DishRepository define las operaciones de persistencia para platos
type DishRepository interface {
	Create(ctx context.Context, dish *Dish) error
	GetByID(ctx context.Context, id uuid.UUID) (*Dish, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]Dish, error)
	GetActiveByCompanyID(ctx context.Context, companyID uuid.UUID) ([]Dish, error)
	GetAvailableForFeed(ctx context.Context, params FeedParams) ([]Dish, int64, error)
	Update(ctx context.Context, dish *Dish) error
	Delete(ctx context.Context, id uuid.UUID) error
	ToggleActive(ctx context.Context, id uuid.UUID) (*Dish, error)
}

// SwipeSessionRepository define las operaciones de persistencia para sesiones de swipe
type SwipeSessionRepository interface {
	Create(ctx context.Context, session *SwipeSession) error
	CreateWithTx(ctx context.Context, tx interface{}, session *SwipeSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*SwipeSession, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]SwipeSession, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*SwipeSession, error)
	GetActiveByUserIDWithTx(ctx context.Context, tx interface{}, userID uuid.UUID) (*SwipeSession, error)
	Update(ctx context.Context, session *SwipeSession) error
	Complete(ctx context.Context, id uuid.UUID) error
	CompleteWithTx(ctx context.Context, tx interface{}, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SwipeActionRepository define las operaciones de persistencia para acciones de swipe
type SwipeActionRepository interface {
	Create(ctx context.Context, action *SwipeAction) error
	GetByID(ctx context.Context, id uuid.UUID) (*SwipeAction, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]SwipeAction, error)
	GetBySessionAndDish(ctx context.Context, sessionID, dishID uuid.UUID) (*SwipeAction, error)
	CountLikesBySession(ctx context.Context, sessionID uuid.UUID) (int, error)
	GetLikesBySession(ctx context.Context, sessionID uuid.UUID) ([]SwipeAction, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// MatchRepository define las operaciones de persistencia para matches
type MatchRepository interface {
	Create(ctx context.Context, match *Match) error
	GetByID(ctx context.Context, id uuid.UUID) (*Match, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]Match, error)
	GetByDishID(ctx context.Context, dishID uuid.UUID) ([]Match, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// AuthService define las operaciones de autenticación
type AuthService interface {
	RegisterUser(ctx context.Context, name, email, password, phone string) (*User, error)
	Login(ctx context.Context, email, password string) (*User, string, error)
	RegisterCompany(ctx context.Context, name, email, password, phone, address string) (*Company, error)
	LoginCompany(ctx context.Context, email, password string) (*Company, string, error)
}

// DishService define las operaciones de negocio para platos
type DishService interface {
	Create(ctx context.Context, companyID uuid.UUID, name, description string, price float64, photoURL string) (*Dish, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Dish, error)
	GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]Dish, error)
	GetActiveByCompanyID(ctx context.Context, companyID uuid.UUID) ([]Dish, error)
	Update(ctx context.Context, id uuid.UUID, name, description string, price float64, photoURL string) (*Dish, error)
	Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error
	ToggleActive(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*Dish, error)
}

// SwipeSessionService define las operaciones de negocio para sesiones de swipe
type SwipeSessionService interface {
	Create(ctx context.Context, userID uuid.UUID, userLat float64, userLng float64, radiusKm int) (*SwipeSession, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SwipeSession, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*SwipeSession, error)
	Complete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetFeed(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, page, perPage int) ([]Dish, int64, error)
	Swipe(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, dishID uuid.UUID, direction SwipeDirection) (*SwipeResult, error)
	GetFinalists(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*FinalistsResult, error)
	Match(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, dishID uuid.UUID) (*MatchResult, error)
}

type MatchResult struct {
	MatchID   uuid.UUID
	DishID    uuid.UUID
	MatchedAt time.Time
	Dish      *Dish
	Company   *Company
}

type SwipeResult struct {
	SwipeID     uuid.UUID
	LikesCount  int
	FeedBlocked bool
}

type FinalistsResult struct {
	SessionID uuid.UUID
	Finalists []FinalistDish
	Count     int
}

type FinalistDish struct {
	ID          string
	Name        string
	Description string
	Price       float64
	PhotoURL    string
	Company     CompanyInfo
	SwipedAt    string
}

type CompanyInfo struct {
	ID      string
	Name    string
	Address string
	Lat     float64
	Lng     float64
	LogoURL string
}

type FeedParams struct {
	SessionID uuid.UUID
	UserLat   float64
	UserLng   float64
	RadiusKm  int
	Limit     int
	Offset    int
}
