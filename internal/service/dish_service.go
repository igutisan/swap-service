package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"swap/iguti/swap-service/internal/domain"
)

type dishService struct {
	dishRepo    domain.DishRepository
	companyRepo domain.CompanyRepository
}

// NewDishService crea una nueva instancia de DishService
func NewDishService(dishRepo domain.DishRepository, companyRepo domain.CompanyRepository) domain.DishService {
	return &dishService{
		dishRepo:    dishRepo,
		companyRepo: companyRepo,
	}
}

func (s *dishService) Create(ctx context.Context, companyID uuid.UUID, name, description string, price float64, photoURL string) (*domain.Dish, error) {
	// Verificar que la compañía exista
	_, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("compañía no encontrada: %w", err)
	}

	// Crear el plato
	dish := &domain.Dish{
		CompanyID:   companyID,
		Name:        name,
		Description: description,
		Price:       price,
		PhotoURL:    photoURL,
		IsActive:    true, // Por defecto activo
	}

	// Guardar en la base de datos
	if err := s.dishRepo.Create(ctx, dish); err != nil {
		return nil, fmt.Errorf("error al crear plato: %w", err)
	}

	return dish, nil
}

func (s *dishService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Dish, error) {
	dish, err := s.dishRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("plato no encontrado: %w", err)
	}

	return dish, nil
}

func (s *dishService) GetByCompanyID(ctx context.Context, companyID uuid.UUID) ([]domain.Dish, error) {
	// Verificar que la compañía exista
	_, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("compañía no encontrada: %w", err)
	}

	dishes, err := s.dishRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener platos: %w", err)
	}

	return dishes, nil
}

func (s *dishService) GetActiveByCompanyID(ctx context.Context, companyID uuid.UUID) ([]domain.Dish, error) {
	// Verificar que la compañía exista
	_, err := s.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("compañía no encontrada: %w", err)
	}

	dishes, err := s.dishRepo.GetActiveByCompanyID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener platos activos: %w", err)
	}

	return dishes, nil
}

func (s *dishService) Update(ctx context.Context, id uuid.UUID, name, description string, price float64, photoURL string) (*domain.Dish, error) {
	// Obtener el plato existente
	dish, err := s.dishRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("plato no encontrado: %w", err)
	}

	// Actualizar campos si se proporcionan
	if name != "" {
		dish.Name = name
	}
	if description != "" {
		dish.Description = description
	}
	if price > 0 {
		dish.Price = price
	}
	if photoURL != "" {
		dish.PhotoURL = photoURL
	}

	// Guardar cambios
	if err := s.dishRepo.Update(ctx, dish); err != nil {
		return nil, fmt.Errorf("error al actualizar plato: %w", err)
	}

	return dish, nil
}

func (s *dishService) Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error {
	// Obtener el plato para verificar que pertenece a la compañía
	dish, err := s.dishRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("plato no encontrado: %w", err)
	}

	// Verificar que el plato pertenece a la compañía
	if dish.CompanyID != companyID {
		return fmt.Errorf("no tienes permiso para eliminar este plato")
	}

	// Eliminar el plato
	if err := s.dishRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("error al eliminar plato: %w", err)
	}

	return nil
}

func (s *dishService) ToggleActive(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*domain.Dish, error) {
	// Obtener el plato para verificar que pertenece a la compañía
	dish, err := s.dishRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("plato no encontrado: %w", err)
	}

	// Verificar que el plato pertenece a la compañía
	if dish.CompanyID != companyID {
		return nil, fmt.Errorf("no tienes permiso para modificar este plato")
	}

	// Toggle del estado
	updatedDish, err := s.dishRepo.ToggleActive(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error al cambiar estado del plato: %w", err)
	}

	return updatedDish, nil
}
