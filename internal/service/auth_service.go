package service

import (
	"context"
	"fmt"
	"strings"

	"swap/iguti/swap-service/internal/domain"
	"swap/iguti/swap-service/internal/utils"
)

type authService struct {
	userRepo         domain.UserRepository
	companyRepo      domain.CompanyRepository
	geocodingService GeocodingService
}

func NewAuthService(userRepo domain.UserRepository, companyRepo domain.CompanyRepository, geocodingService GeocodingService) domain.AuthService {
	return &authService{
		userRepo:         userRepo,
		companyRepo:      companyRepo,
		geocodingService: geocodingService,
	}
}

func (s *authService) RegisterUser(ctx context.Context, name, email, password, phone string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrEmailAlreadyExist
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error al procesar la contraseña: %w", err)
	}

	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Phone:    phone,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("error al crear usuario: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, string, error) {

	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", domain.ErrUserNotFound
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", domain.ErrInvalidPassword
	}

	token, err := utils.GenerateAuthJWT(user.ID, user.Email, utils.ActorTypeUser)
	if err != nil {
		return nil, "", fmt.Errorf("error al generar token: %w", err)
	}

	return user, token, nil
}

// RegisterCompany registra una nueva empresa en el sistema
func (s *authService) RegisterCompany(ctx context.Context, name, email, password, phone, address string) (*domain.Company, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	existingCompany, err := s.companyRepo.GetByEmail(ctx, email)
	if err == nil && existingCompany != nil {
		return nil, domain.ErrEmailAlreadyExist
	}

	geocodeResult, err := s.geocodingService.Geocode(ctx, address)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("error al procesar la contraseña: %w", err)
	}

	company := &domain.Company{
		Name:     name,
		Email:    email,
		Password: hashedPassword,
		Phone:    phone,
		Address:  address,
		Lat:      geocodeResult.Lat,
		Lng:      geocodeResult.Lng,
	}

	if err := s.companyRepo.Create(ctx, company); err != nil {
		return nil, fmt.Errorf("error al crear empresa: %w", err)
	}

	return company, nil
}

// LoginCompany autentica una empresa y retorna un token JWT
func (s *authService) LoginCompany(ctx context.Context, email, password string) (*domain.Company, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	company, err := s.companyRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", domain.ErrCompanyNotFound
	}

	if !utils.CheckPasswordHash(password, company.Password) {
		return nil, "", domain.ErrInvalidPassword
	}

	token, err := utils.GenerateAuthJWT(company.ID, company.Email, utils.ActorTypeCompany)
	if err != nil {
		return nil, "", fmt.Errorf("error al generar token: %w", err)
	}

	return company, token, nil
}
