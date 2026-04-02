package dto

import (
	"github.com/google/uuid"
)

// ==================== REQUESTS ====================

// CreateDishRequest representa la petición para crear un plato
type CreateDishRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description string  `json:"description" validate:"omitempty,max=500"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	PhotoURL    string  `json:"photo_url" validate:"omitempty,url,max=500"`
}

// UpdateDishRequest representa la petición para actualizar un plato
type UpdateDishRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=2,max=100"`
	Description string  `json:"description" validate:"omitempty,max=500"`
	Price       float64 `json:"price" validate:"omitempty,gt=0"`
	PhotoURL    string  `json:"photo_url" validate:"omitempty,url,max=500"`
}

// DishIDRequest representa el parámetro de ID del plato en la URL
type DishIDRequest struct {
	ID string `uri:"id" validate:"required,uuid"`
}

// ==================== DATA RESPONSES ====================

// DishData representa los datos de un plato en la respuesta
type DishData struct {
	ID          uuid.UUID `json:"id"`
	CompanyID   uuid.UUID `json:"company_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	PhotoURL    string    `json:"photo_url,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at,omitempty"`
}

// DishListData representa una lista de platos
type DishListData struct {
	Dishes []DishData `json:"dishes"`
	Count  int        `json:"count"`
}

// ToggleDishStatusData representa los datos después de cambiar el estado
type ToggleDishStatusData struct {
	ID       uuid.UUID `json:"id"`
	IsActive bool      `json:"is_active"`
	Status   string    `json:"status"`
}
