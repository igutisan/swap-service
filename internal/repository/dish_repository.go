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

	// Debug: Check total dishes in DB
	var dbTotal int64
	r.db.WithContext(ctx).Model(&domain.Dish{}).Count(&dbTotal)
	var dbActive int64
	r.db.WithContext(ctx).Model(&domain.Dish{}).Where("is_active = ?", true).Count(&dbActive)
	log.Printf("[DEBUG GetAvailableForFeed] Total dishes in DB: %d, Active: %d", dbTotal, dbActive)

	// Get swiped dish IDs
	var swipedIDs []uuid.UUID
	r.db.WithContext(ctx).
		Table("swipe_actions").
		Where("session_id = ?", params.SessionID).
		Pluck("dish_id", &swipedIDs)

	query := r.db.WithContext(ctx).
		Model(&domain.Dish{}).
		Table("dishes d").
		Joins("INNER JOIN companies c ON d.company_id = c.id").
		Where("d.is_active = ?", true)

	if len(swipedIDs) > 0 {
		query = query.Where("d.id NOT IN (?)", swipedIDs)
	}

	spatialFilter := fmt.Sprintf(`
		ST_DWithin(
			ST_SetSRID(ST_MakePoint(c.lng, c.lat), 4326)::geography,
			ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography,
			%f
		)
	`, userLng, userLat, radiusMeters)

	query = query.Where(spatialFilter)

	// Count total available
	if err := query.Count(&total).Error; err != nil {
		log.Printf("[ERROR GetAvailableForFeed] Count error: %v", err)
		return nil, 0, err
	}

	// Fetch page
	err := query.
		Order("d.created_at DESC").
		Limit(params.Limit).
		Offset(params.Offset).
		Preload("Company").
		Find(&dishes).Error

	if err != nil {
		return nil, 0, fmt.Errorf("error al obtener platos disponibles: %w", err)
	}

	log.Printf("[DEBUG dishRepository] GetAvailableForFeed: session=%s lat=%f lng=%f radius=%dm -> found %d/%d dishes (swiped=%d)",
		params.SessionID, userLat, userLng, int(radiusMeters), len(dishes), total, len(swipedIDs))

	// Debug: If total is 0, let's see distances of all active dishes
	if total == 0 && dbActive > 0 {
		var allActive []domain.Dish
		r.db.WithContext(ctx).Preload("Company").Where("is_active = ?", true).Find(&allActive)
		for _, ad := range allActive {
			log.Printf("[DEBUG GetAvailableForFeed] Active Dish: %s, Company: %s, Lat: %f, Lng: %f", ad.Name, ad.Company.Name, ad.Company.Lat, ad.Company.Lng)
		}
	}

	return dishes, total, nil
}
