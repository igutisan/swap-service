package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"swap/iguti/swap-service/internal/domain"
)

type companyRepository struct {
	db *gorm.DB
}

// NewCompanyRepository crea una nueva instancia de CompanyRepository
func NewCompanyRepository(db *gorm.DB) domain.CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(ctx context.Context, company *domain.Company) error {
	return r.db.WithContext(ctx).Create(company).Error
}

func (r *companyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	var company domain.Company
	err := r.db.WithContext(ctx).First(&company, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) GetByEmail(ctx context.Context, email string) (*domain.Company, error) {
	var company domain.Company
	err := r.db.WithContext(ctx).First(&company, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) Update(ctx context.Context, company *domain.Company) error {
	return r.db.WithContext(ctx).Save(company).Error
}

func (r *companyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Company{}, "id = ?", id).Error
}
