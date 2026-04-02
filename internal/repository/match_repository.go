package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"swap/iguti/swap-service/internal/domain"
)

type matchRepository struct {
	db *gorm.DB
}

// NewMatchRepository crea una nueva instancia de MatchRepository
func NewMatchRepository(db *gorm.DB) domain.MatchRepository {
	return &matchRepository{db: db}
}

func (r *matchRepository) Create(ctx context.Context, match *domain.Match) error {
	return r.db.WithContext(ctx).Create(match).Error
}

func (r *matchRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Match, error) {
	var match domain.Match
	err := r.db.WithContext(ctx).First(&match, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &match, nil
}

func (r *matchRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]domain.Match, error) {
	var matches []domain.Match
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Find(&matches).Error
	return matches, err
}

func (r *matchRepository) GetByDishID(ctx context.Context, dishID uuid.UUID) ([]domain.Match, error) {
	var matches []domain.Match
	err := r.db.WithContext(ctx).Where("dish_id = ?", dishID).Find(&matches).Error
	return matches, err
}

func (r *matchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Match{}, "id = ?", id).Error
}
