package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"swap/iguti/swap-service/internal/domain"
)

type swipeActionRepository struct {
	db *gorm.DB
}

// NewSwipeActionRepository crea una nueva instancia de SwipeActionRepository
func NewSwipeActionRepository(db *gorm.DB) domain.SwipeActionRepository {
	return &swipeActionRepository{db: db}
}

func (r *swipeActionRepository) Create(ctx context.Context, action *domain.SwipeAction) error {
	return r.db.WithContext(ctx).Create(action).Error
}

func (r *swipeActionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SwipeAction, error) {
	var action domain.SwipeAction
	err := r.db.WithContext(ctx).First(&action, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &action, nil
}

func (r *swipeActionRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]domain.SwipeAction, error) {
	var actions []domain.SwipeAction
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Find(&actions).Error
	return actions, err
}

func (r *swipeActionRepository) GetBySessionAndDish(ctx context.Context, sessionID, dishID uuid.UUID) (*domain.SwipeAction, error) {
	var action domain.SwipeAction
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND dish_id = ?", sessionID, dishID).
		First(&action).Error
	if err != nil {
		return nil, err
	}
	return &action, nil
}

func (r *swipeActionRepository) CountLikesBySession(ctx context.Context, sessionID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.SwipeAction{}).
		Where("session_id = ? AND direction = ?", sessionID, domain.SwipeDirectionRight).
		Count(&count).Error
	return int(count), err
}

func (r *swipeActionRepository) GetLikesBySession(ctx context.Context, sessionID uuid.UUID) ([]domain.SwipeAction, error) {
	var actions []domain.SwipeAction
	err := r.db.WithContext(ctx).
		Preload("Dish.Company").
		Where("session_id = ? AND direction = ?", sessionID, domain.SwipeDirectionRight).
		Order("swiped_at DESC").
		Find(&actions).Error
	return actions, err
}

func (r *swipeActionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.SwipeAction{}, "id = ?", id).Error
}
