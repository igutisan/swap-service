package domain

import (
	"time"

	"github.com/google/uuid"
)

type Dish struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	PhotoURL    string    `gorm:"type:varchar(500)" json:"photo_url"`
	IsActive    bool      `gorm:"type:boolean;not null;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"type:timestamp;not null;default:now()" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp;not null;autoUpdateTime" json:"updated_at"`

	// Relaciones
	Company      Company       `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	SwipeActions []SwipeAction `gorm:"foreignKey:DishID" json:"swipe_actions,omitempty"`
	Matches      []Match       `gorm:"foreignKey:DishID" json:"matches,omitempty"`
}

func (Dish) TableName() string {
	return "dishes"
}
