package dto

import (
	"github.com/google/uuid"
)

// ==================== REQUESTS ====================

// CreateSessionRequest representa la petición para crear una sesión de swipe
type CreateSessionRequest struct {
	UserLat  float64 `json:"user_lat" validate:"required,latitude,not_zero"`
	UserLng  float64 `json:"user_lng" validate:"required,longitude,not_zero"`
	RadiusKm int     `json:"radius_km" validate:"required,min=1,max=50"`
}

// SessionIDParam representa el parámetro de ID de sesión en la URL
type SessionIDParam struct {
	ID string `uri:"id" validate:"required,uuid"`
}

// FeedPaginationRequest incluye parámetros de paginación para el feed
type FeedPaginationRequest struct {
	Page    int `form:"page" json:"page" validate:"omitempty,min=1"`
	PerPage int `form:"per_page" json:"per_page" validate:"omitempty,min=1,max=50"`
}

// GetOrDefault retorna valores por defecto si no están establecidos
func (f FeedPaginationRequest) GetOrDefault() (page, perPage int) {
	page = f.Page
	perPage = f.PerPage

	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	return page, perPage
}

// ==================== DATA RESPONSES ====================

// SessionData representa los datos de una sesión de swipe
type SessionData struct {
	ID          uuid.UUID `json:"session_id"`
	UserID      uuid.UUID `json:"user_id,omitempty"`
	UserLat     float64   `json:"user_lat"`
	UserLng     float64   `json:"user_lng"`
	RadiusKm    int       `json:"radius_km"`
	Status      string    `json:"status"`
	StartedAt   string    `json:"started_at"`
	CompletedAt *string   `json:"completed_at,omitempty"`
}

// FeedCompanyInfo representa la información de la compañía en el feed
type FeedCompanyInfo struct {
	Name       string  `json:"name"`
	DistanceKm float64 `json:"distance_km"`
	Address    string  `json:"address,omitempty"`
	LogoURL    string  `json:"logo_url,omitempty"`
}

// FeedDishItem representa un plato en el feed de swipe
type FeedDishItem struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Price       float64         `json:"price"`
	PhotoURL    string          `json:"photo_url,omitempty"`
	Company     FeedCompanyInfo `json:"company"`
}

// FeedData representa el feed de platos disponibles para swipe con paginación
type FeedData struct {
	SessionID uuid.UUID      `json:"session_id"`
	UserLat   float64        `json:"user_lat"`
	UserLng   float64        `json:"user_lng"`
	RadiusKm  int            `json:"radius_km"`
	Dishes    []FeedDishItem `json:"dishes"`
}

// SwipeRequest representa la petición para hacer swipe
type SwipeRequest struct {
	DishID    string `json:"dish_id" validate:"required,uuid"`
	Direction string `json:"direction" validate:"required,oneof=left right"`
}

// SwipeResponse representa la respuesta después de hacer swipe
type SwipeResponse struct {
	SwipeID     string `json:"swipe_id"`
	LikesCount  int    `json:"likes_count"`
	FeedBlocked bool   `json:"feed_blocked"`
}

// FinalistDish representa un plato finalista (like)
type FinalistDish struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Price       float64         `json:"price"`
	PhotoURL    string          `json:"photo_url"`
	Company     FinalistCompany `json:"company"`
	SwipedAt    string          `json:"swiped_at"`
}

// FinalistCompany representa la información de la empresa en finalistas
type FinalistCompany struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	LogoURL string  `json:"logo_url,omitempty"`
}

// FinalistsResponse representa la respuesta de los finalistas
type FinalistsResponse struct {
	SessionID string         `json:"session_id"`
	Finalists []FinalistDish `json:"finalists"`
	Count     int            `json:"count"`
}

// MatchRequest representa la petición para crear un match
type MatchRequest struct {
	DishID string `json:"dish_id" validate:"required,uuid"`
}

// MatchDish representa la información del dish en el match
type MatchDish struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	PhotoURL string  `json:"photo_url"`
}

// MatchCompany representa la información de la empresa en el match
type MatchCompany struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Phone   string  `json:"phone"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// MatchResponse representa la respuesta después de crear un match
type MatchResponse struct {
	Dish      MatchDish    `json:"dish"`
	Company   MatchCompany `json:"company"`
	MatchedAt string       `json:"matched_at"`
}
