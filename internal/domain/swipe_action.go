package domain

import (
	"time"

	"github.com/google/uuid"
)

type SwipeDirection string

const (
	SwipeDirectionLeft  SwipeDirection = "left"
	SwipeDirectionRight SwipeDirection = "right"
)

type SwipeAction struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SessionID uuid.UUID      `gorm:"type:uuid;not null;index" json:"session_id"`
	DishID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"dish_id"`
	Direction SwipeDirection `gorm:"type:varchar(10);not null" json:"direction"`
	SwipedAt  time.Time      `gorm:"type:timestamp;not null;default:now()" json:"swiped_at"`

	// Relaciones
	Session SwipeSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Dish    Dish         `gorm:"foreignKey:DishID" json:"dish,omitempty"`
}

func (SwipeAction) TableName() string {
	return "swipe_actions"
}
