package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"swap/iguti/swap-service/internal/domain"
)

type swipeSessionService struct {
	sessionRepo domain.SwipeSessionRepository
	dishRepo    domain.DishRepository
	actionRepo  domain.SwipeActionRepository
	matchRepo   domain.MatchRepository
	txManager   interface{}
}

func NewSwipeSessionService(
	sessionRepo domain.SwipeSessionRepository,
	dishRepo domain.DishRepository,
	actionRepo domain.SwipeActionRepository,
	matchRepo domain.MatchRepository,
	txManager interface{},
) domain.SwipeSessionService {
	return &swipeSessionService{
		sessionRepo: sessionRepo,
		dishRepo:    dishRepo,
		actionRepo:  actionRepo,
		matchRepo:   matchRepo,
		txManager:   txManager,
	}
}

func (s *swipeSessionService) Create(ctx context.Context, userID uuid.UUID, userLat float64, userLng float64, radiusKm int) (*domain.SwipeSession, error) {
	if userLat == 0 && userLng == 0 {
		return nil, fmt.Errorf("coordenadas inválidas: latitud y longitud no pueden ser ambas cero")
	}

	if radiusKm < domain.MinRadiusKm || radiusKm > domain.MaxRadiusKm {
		return nil, fmt.Errorf("el radio debe estar entre %d y %d km", domain.MinRadiusKm, domain.MaxRadiusKm)
	}

	tx, err := s.txManager.(domain.TransactionManager).BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			s.txManager.(domain.TransactionManager).Rollback(tx)
		}
	}()

	existingSession, err := s.sessionRepo.GetActiveByUserIDWithTx(ctx, tx, userID)
	if err == nil && existingSession != nil {
		if err := s.sessionRepo.CompleteWithTx(ctx, tx, existingSession.ID); err != nil {
			s.txManager.(domain.TransactionManager).Rollback(tx)
			return nil, fmt.Errorf("error al completar sesión anterior: %w", err)
		}
	}

	session := &domain.SwipeSession{
		UserID:   userID,
		UserLat:  userLat,
		UserLng:  userLng,
		RadiusKm: radiusKm,
		Status:   domain.SessionStatusActive,
	}

	if err := s.sessionRepo.CreateWithTx(ctx, tx, session); err != nil {
		s.txManager.(domain.TransactionManager).Rollback(tx)
		return nil, fmt.Errorf("error al crear sesión: %w", err)
	}

	if err := s.txManager.(domain.TransactionManager).Commit(tx); err != nil {
		return nil, fmt.Errorf("error al confirmar transacción: %w", err)
	}

	return session, nil
}

func (s *swipeSessionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.SwipeSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sesión no encontrada: %w", err)
	}
	return session, nil
}

func (s *swipeSessionService) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*domain.SwipeSession, error) {
	session, err := s.sessionRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no hay sesión activa: %w", err)
	}
	return session, nil
}

func (s *swipeSessionService) Complete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("sesión no encontrada: %w", err)
	}

	if err := s.validateSessionOwnership(session, userID); err != nil {
		return err
	}

	if err := s.validateSessionActive(session); err != nil {
		return err
	}

	if err := s.sessionRepo.Complete(ctx, id); err != nil {
		return fmt.Errorf("error al completar sesión: %w", err)
	}

	return nil
}

func (s *swipeSessionService) GetFeed(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, page, perPage int) ([]domain.Dish, int64, error) {
	session, err := s.getAndValidateSession(ctx, sessionID, userID)
	if err != nil {
		return nil, 0, err
	}

	log.Printf("[DEBUG GetFeed] SessionID=%s UserLat=%f UserLng=%f RadiusKm=%d",
		sessionID, session.UserLat, session.UserLng, session.RadiusKm)

	page, perPage = s.validateFeedParams(page, perPage)

	dishes, total, err := s.getAvailableDishes(ctx, session, page, perPage)
	if err != nil {
		return nil, 0, err
	}

	log.Printf("[DEBUG GetFeed] Found %d dishes (total: %d)", len(dishes), total)
	for i, dish := range dishes {
		log.Printf("[DEBUG GetFeed] Dish[%d] ID=%s Company=%s CompanyLat=%f CompanyLng=%f",
			i, dish.ID, dish.Company.Name, dish.Company.Lat, dish.Company.Lng)
	}

	return dishes, total, nil
}

func (s *swipeSessionService) validateSessionOwnership(session *domain.SwipeSession, userID uuid.UUID) error {
	if session.UserID != userID {
		return fmt.Errorf("no tienes permiso para acceder a esta sesión")
	}
	return nil
}

func (s *swipeSessionService) validateSessionActive(session *domain.SwipeSession) error {
	if session.Status != domain.SessionStatusActive {
		return fmt.Errorf("la sesión no está activa")
	}
	return nil
}

func (s *swipeSessionService) getAndValidateSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*domain.SwipeSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("sesión no encontrada: %w", err)
	}

	if err := s.validateSessionOwnership(session, userID); err != nil {
		return nil, err
	}

	if err := s.validateSessionActive(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *swipeSessionService) validateFeedParams(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}

	if perPage < domain.MinPerPage {
		perPage = domain.MinPerPage
	}
	if perPage > domain.MaxPerPage {
		perPage = domain.MaxPerPage
	}

	return page, perPage
}

func (s *swipeSessionService) getAvailableDishes(ctx context.Context, session *domain.SwipeSession, page, perPage int) ([]domain.Dish, int64, error) {
	params := domain.FeedParams{
		SessionID: session.ID,
		UserLat:   session.UserLat,
		UserLng:   session.UserLng,
		RadiusKm:  session.RadiusKm,
		Limit:     perPage,
		Offset:    (page - 1) * perPage,
	}

	return s.dishRepo.GetAvailableForFeed(ctx, params)
}

func (s *swipeSessionService) Swipe(ctx context.Context, sessionID, userID, dishID uuid.UUID, direction domain.SwipeDirection) (*domain.SwipeResult, error) {
	_, err := s.getAndValidateSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	dish, err := s.dishRepo.GetByID(ctx, dishID)
	if err != nil {
		return nil, fmt.Errorf("plato no encontrado: %w", err)
	}

	if !dish.IsActive {
		return nil, fmt.Errorf("el plato no está disponible")
	}

	_, err = s.actionRepo.GetBySessionAndDish(ctx, sessionID, dishID)
	if err == nil {
		return nil, domain.ErrDishAlreadySwiped
	}

	if direction == domain.SwipeDirectionRight {
		currentLikes, err := s.actionRepo.CountLikesBySession(ctx, sessionID)
		if err != nil {
			return nil, fmt.Errorf("error al contar likes: %w", err)
		}

		if currentLikes >= domain.MaxLikesPerSession {
			return nil, domain.ErrMaxLikesReached
		}
	}

	action := &domain.SwipeAction{
		SessionID: sessionID,
		DishID:    dishID,
		Direction: direction,
	}

	if err := s.actionRepo.Create(ctx, action); err != nil {
		return nil, fmt.Errorf("error al registrar swipe: %w", err)
	}

	totalLikes, err := s.actionRepo.CountLikesBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("error al contar likes: %w", err)
	}

	feedBlocked := totalLikes >= domain.MaxLikesPerSession

	return &domain.SwipeResult{
		SwipeID:     action.ID,
		LikesCount:  totalLikes,
		FeedBlocked: feedBlocked,
	}, nil
}

func (s *swipeSessionService) GetFinalists(ctx context.Context, sessionID, userID uuid.UUID) (*domain.FinalistsResult, error) {
	_, err := s.getAndValidateSession(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	likes, err := s.actionRepo.GetLikesBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener finalistas: %w", err)
	}

	var finalists []domain.FinalistDish
	for _, action := range likes {
		finalists = append(finalists, domain.FinalistDish{
			ID:          action.Dish.ID.String(),
			Name:        action.Dish.Name,
			Description: action.Dish.Description,
			Price:       action.Dish.Price,
			PhotoURL:    action.Dish.PhotoURL,
			Company: domain.CompanyInfo{
				ID:      action.Dish.Company.ID.String(),
				Name:    action.Dish.Company.Name,
				Address: action.Dish.Company.Address,
				Lat:     action.Dish.Company.Lat,
				Lng:     action.Dish.Company.Lng,
				LogoURL: action.Dish.Company.LogoURL,
			},
			SwipedAt: action.SwipedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &domain.FinalistsResult{
		SessionID: sessionID,
		Finalists: finalists,
		Count:     len(finalists),
	}, nil
}

func (s *swipeSessionService) Match(ctx context.Context, sessionID, userID, dishID uuid.UUID) (*domain.MatchResult, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("sesión no encontrada: %w", err)
	}

	if session.UserID != userID {
		return nil, domain.ErrUnauthorizedSession
	}

	if session.Status == domain.SessionStatusCompleted {
		return nil, domain.ErrSessionAlreadyCompleted
	}

	if session.Status != domain.SessionStatusActive {
		return nil, domain.ErrSessionNotActive
	}

	likes, err := s.actionRepo.GetLikesBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener finalistas: %w", err)
	}

	isFinalist := false
	for _, action := range likes {
		if action.DishID == dishID {
			isFinalist = true
			break
		}
	}

	if !isFinalist {
		return nil, domain.ErrDishNotInFinalists
	}

	dish, err := s.dishRepo.GetByID(ctx, dishID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener plato: %w", err)
	}

	match := &domain.Match{
		SessionID: sessionID,
		DishID:    dishID,
	}

	if err := s.matchRepo.Create(ctx, match); err != nil {
		return nil, fmt.Errorf("error al crear match: %w", err)
	}

	if err := s.sessionRepo.Complete(ctx, sessionID); err != nil {
		return nil, fmt.Errorf("error al completar sesión: %w", err)
	}

	return &domain.MatchResult{
		MatchID:   match.ID,
		DishID:    match.DishID,
		MatchedAt: match.MatchedAt,
		Dish:      dish,
		Company:   &dish.Company,
	}, nil
}
