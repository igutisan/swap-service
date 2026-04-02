package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	Email        string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	Password     string    `gorm:"type:varchar(255);not null" json:"-"`
	Phone        string    `gorm:"type:varchar(50)" json:"phone"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:now()" json:"created_at"`

	// Relaciones
	SwipeSessions []SwipeSession `gorm:"foreignKey:UserID" json:"swipe_sessions,omitempty"`
}

func (User) TableName() string {
	return "users"
}
