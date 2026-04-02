package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"swap/iguti/swap-service/internal/domain"
)

type swipeSessionRepository struct {
	db *gorm.DB
}

// NewSwipeSessionRepository crea una nueva instancia de SwipeSessionRepository
func NewSwipeSessionRepository(db *gorm.DB) domain.SwipeSessionRepository {
	return &swipeSessionRepository{db: db}
}

func (r *swipeSessionRepository) Create(ctx context.Context, session *domain.SwipeSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *swipeSessionRepository) CreateWithTx(ctx context.Context, tx interface{}, session *domain.SwipeSession) error {
	return tx.(*gorm.DB).WithContext(ctx).Create(session).Error
}

func (r *swipeSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SwipeSession, error) {
	var session domain.SwipeSession
	err := r.db.WithContext(ctx).First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *swipeSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.SwipeSession, error) {
	var sessions []domain.SwipeSession
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error
	return sessions, err
}

func (r *swipeSessionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*domain.SwipeSession, error) {
	var session domain.SwipeSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.SessionStatusActive).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *swipeSessionRepository) GetActiveByUserIDWithTx(ctx context.Context, tx interface{}, userID uuid.UUID) (*domain.SwipeSession, error) {
	var session domain.SwipeSession
	err := tx.(*gorm.DB).WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, domain.SessionStatusActive).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *swipeSessionRepository) Update(ctx context.Context, session *domain.SwipeSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *swipeSessionRepository) Complete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.SwipeSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       domain.SessionStatusCompleted,
			"completed_at": now,
		}).Error
}

func (r *swipeSessionRepository) CompleteWithTx(ctx context.Context, tx interface{}, id uuid.UUID) error {
	now := time.Now()
	return tx.(*gorm.DB).WithContext(ctx).
		Model(&domain.SwipeSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       domain.SessionStatusCompleted,
			"completed_at": now,
		}).Error
}

func (r *swipeSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.SwipeSession{}, "id = ?", id).Error
}
