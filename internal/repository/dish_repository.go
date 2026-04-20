package repository

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"swap/iguti/swap-service/internal/domain"
)

type dishRepository struct {
	db *gorm.DB
}

// NewDishRepository crea una nueva instancia de DishRepository
func NewDishRepository(db *gorm.DB) domain.DishRepository {
	return &dishRepository{db: db}
}

func (r *dishRepository) Create(ctx context.Context, dish *domain.Dish) error {
	return r.db.WithContext(ctx).Create(dish).Error
}

func (r *dishRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Dish, error) {
	var dish domain.Dish
	err := r.db.WithContext(ctx).First(&dish, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &dish, nil
}

func (r *dishRepository) GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]domain.Dish, error) {
	var dishes []domain.Dish
	err := r.db.WithContext(ctx).Where("company_id = ?", companyID).Find(&dishes).Error
	return dishes, err
}

func (r *dishRepository) GetActiveByCompanyID(ctx context.Context, companyID uuid.UUID) ([]domain.Dish, error) {
	var dishes []domain.Dish
	err := r.db.WithContext(ctx).
		Where("company_id = ? AND is_active = ?", companyID, true).
		Find(&dishes).Error
	return dishes, err
}

func (r *dishRepository) Update(ctx context.Context, dish *domain.Dish) error {
	return r.db.WithContext(ctx).Save(dish).Error
}

func (r *dishRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Dish{}, "id = ?", id).Error
}

func (r *dishRepository) ToggleActive(ctx context.Context, id uuid.UUID) (*domain.Dish, error) {
	var dish domain.Dish
	err := r.db.WithContext(ctx).First(&dish, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	dish.IsActive = !dish.IsActive
	if err := r.db.WithContext(ctx).Save(&dish).Error; err != nil {
		return nil, err
	}

	return &dish, nil
}

func (r *dishRepository) GetAvailableForFeed(ctx context.Context, params domain.FeedParams) ([]domain.Dish, int64, error) {
	var dishes []domain.Dish
	var total int64

	userLng := params.UserLng
	userLat := params.UserLat
	radiusMeters := float64(params.RadiusKm) * 1000

	subquerySwiped := r.db.WithContext(ctx).
		Table("swipe_actions").
		Select("dish_id").
		Where("session_id = ?", params.SessionID)

	r.db.WithContext(ctx).
		Model(&domain.Dish{}).
		Table("dishes d").
		Joins("INNER JOIN companies c ON d.company_id = c.id").
		Where("d.is_active = ?", true).
		Where("d.id NOT IN (?)", subquerySwiped).
		Where(fmt.Sprintf(`
			ST_DWithin(
				ST_MakePoint(c.lng, c.lat)::geography,
				ST_MakePoint(%f, %f)::geography,
				%f
			)
		`, userLng, userLat, radiusMeters)).
		Count(&total)

	err := r.db.WithContext(ctx).
		Model(&domain.Dish{}).
		Table("dishes d").
		Joins("INNER JOIN companies c ON d.company_id = c.id").
		Where("d.is_active = ?", true).
		Where("d.id NOT IN (?)", subquerySwiped).
		Where(fmt.Sprintf(`
			ST_DWithin(
				ST_MakePoint(c.lng, c.lat)::geography,
				ST_MakePoint(%f, %f)::geography,
				%f
			)
		`, userLng, userLat, radiusMeters)).
		Order("d.created_at DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Preload("Company").
		Find(&dishes).Error

	if err != nil {
		return nil, 0, fmt.Errorf("error al obtener platos disponibles: %w", err)
	}

	log.Printf("[DEBUG dishRepository] GetAvailableForFeed: session=%s lat=%f lng=%f radius=%dm -> found %d/%d dishes",
		params.SessionID, userLat, userLng, int(radiusMeters), len(dishes), total)

	return dishes, total, nil
}
