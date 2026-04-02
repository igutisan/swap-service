package domain

import (
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
)

type SwipeSession struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID     `gorm:"type:uuid;not null;index" json:"user_id"`
	UserLat     float64       `gorm:"type:decimal(10,7);not null" json:"user_lat"`
	UserLng     float64       `gorm:"type:decimal(10,7);not null" json:"user_lng"`
	RadiusKm    int           `gorm:"type:int;not null" json:"radius_km"`
	Status      SessionStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	StartedAt   time.Time     `gorm:"type:timestamp;not null;default:now()" json:"started_at"`
	CompletedAt *time.Time    `gorm:"type:timestamp" json:"completed_at,omitempty"`

	// Relaciones
	User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SwipeActions []SwipeAction `gorm:"foreignKey:SessionID" json:"swipe_actions,omitempty"`
	Matches      []Match       `gorm:"foreignKey:SessionID" json:"matches,omitempty"`
}

func (SwipeSession) TableName() string {
	return "swipe_sessions"
}
