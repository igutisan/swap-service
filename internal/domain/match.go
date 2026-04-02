package domain

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SessionID uuid.UUID `gorm:"type:uuid;not null;index" json:"session_id"`
	DishID    uuid.UUID `gorm:"type:uuid;not null;index" json:"dish_id"`
	MatchedAt time.Time `gorm:"type:timestamp;not null;default:now()" json:"matched_at"`

	// Relaciones
	Session SwipeSession `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Dish    Dish         `gorm:"foreignKey:DishID" json:"dish,omitempty"`
}

func (Match) TableName() string {
	return "matches"
}
