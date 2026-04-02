package domain

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	Password string    `gorm:"type:varchar(255);not null" json:"-"`
	Phone        string    `gorm:"type:varchar(50)" json:"phone"`
	Address      string    `gorm:"type:varchar(500)" json:"address"`
	Lat          float64   `gorm:"type:decimal(10,7)" json:"lat"`
	Lng          float64   `gorm:"type:decimal(10,7)" json:"lng"`
	LogoURL      string    `gorm:"type:varchar(500)" json:"logo_url"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:now()" json:"created_at"`

	// Relaciones
	Dishes []Dish `gorm:"foreignKey:CompanyID" json:"dishes,omitempty"`
}

func (Company) TableName() string {
	return "companies"
}
