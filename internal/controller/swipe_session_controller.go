package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"swap/iguti/swap-service/internal/domain"
	"swap/iguti/swap-service/internal/dto"
	"swap/iguti/swap-service/internal/middleware"
	"swap/iguti/swap-service/internal/utils"
)

// SwipeSessionController maneja los endpoints de sesiones de swipe
type SwipeSessionController struct {
	sessionService domain.SwipeSessionService
}

// NewSwipeSessionController crea una nueva instancia de SwipeSessionController
func NewSwipeSessionController(sessionService domain.SwipeSessionService) *SwipeSessionController {
	return &SwipeSessionController{
		sessionService: sessionService,
	}
}

// Create godoc
// @Summary Crea una nueva sesión de swipe
// @Description Inicia una sesión de swipe con ubicación y radio de búsqueda
// @Tags sessions
// @Accept json
// @Produce json
// @Param request body dto.CreateSessionRequest true "Datos de ubicación y radio"
// @Success 201 {object} dto.APIResponse{data=dto.SessionData}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /sessions [post]
func (c *SwipeSessionController) Create(ctx *gin.Context) {
	// Obtener el userID del contexto (inyectado por el middleware JWT)
	userIDStr := ctx.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "ID de usuario inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Obtener request validado del contexto
	req := middleware.GetValidatedRequest(ctx).(*dto.CreateSessionRequest)

	// Llamar al servicio
	session, err := c.sessionService.Create(ctx, userID, req.UserLat, req.UserLng, req.RadiusKm)
	if err != nil {
		resp := dto.ErrorResponse(400, "VALIDATION_ERROR: "+err.Error())
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// Construir respuesta exitosa
	data := dto.SessionData{
		ID:        session.ID,
		UserID:    session.UserID,
		UserLat:   session.UserLat,
		UserLng:   session.UserLng,
		RadiusKm:  session.RadiusKm,
		Status:    string(session.Status),
		StartedAt: session.StartedAt.Format("2006-01-02T15:04:05Z"),
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Sesión de swipe creada exitosamente"),
	)
	ctx.JSON(http.StatusCreated, resp)
}

// GetFeed godoc
// @Summary Obtiene el feed de platos para una sesión
// @Description Retorna platos activos dentro del radio de búsqueda que aún no han sido swiped
// @Tags sessions
// @Produce json
// @Param id path string true "ID de la sesión (UUID)"
// @Param page query int false "Número de página (default: 1)" default(1)
// @Param per_page query int false "Items por página (default: 20, max: 50)" default(20)
// @Success 200 {object} dto.APIResponse{data=dto.PaginatedData}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /sessions/{id}/feed [get]
func (c *SwipeSessionController) GetFeed(ctx *gin.Context) {
	// Obtener el userID del contexto
	userIDStr := ctx.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "ID de usuario inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Obtener el ID de la sesión de la URL
	sessionIDStr := ctx.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de sesión inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// Obtener parámetros de paginación
	var paginationReq dto.FeedPaginationRequest
	if err := ctx.ShouldBindQuery(&paginationReq); err != nil {
		resp := dto.ErrorResponse(400, "INVALID_PAGINATION: Parámetros de paginación inválidos", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// Establecer valores por defecto
	page, perPage := paginationReq.GetOrDefault()

	// Obtener la sesión para incluir datos en la respuesta
	session, err := c.sessionService.GetByID(ctx, sessionID)
	if err != nil {
		resp := dto.ErrorResponse(404, "NOT_FOUND: Sesión no encontrada", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusNotFound, resp)
		return
	}

	// Obtener el feed con paginación
	dishes, totalItems, err := c.sessionService.GetFeed(ctx, sessionID, userID, page, perPage)
	if err != nil {
		resp := dto.ErrorResponse(500, "INTERNAL_ERROR: Error al obtener feed", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Convertir platos a DTOs con cálculo de distancia
	var feedItems []dto.FeedDishItem
	for _, dish := range dishes {
		// Calcular distancia
		distance := utils.CalculateDistance(
			session.UserLat,
			session.UserLng,
			dish.Company.Lat,
			dish.Company.Lng,
		)
		distance = utils.RoundToDecimals(distance, 1) // Redondear a 1 decimal

		feedItems = append(feedItems, dto.FeedDishItem{
			ID:          dish.ID,
			Name:        dish.Name,
			Description: dish.Description,
			Price:       dish.Price,
			PhotoURL:    dish.PhotoURL,
			Company: dto.FeedCompanyInfo{
				Name:       dish.Company.Name,
				DistanceKm: distance,
				Address:    dish.Company.Address,
				LogoURL:    dish.Company.LogoURL,
			},
		})
	}

	// Construir datos del feed
	feedData := dto.FeedData{
		SessionID: session.ID,
		UserLat:   session.UserLat,
		UserLng:   session.UserLng,
		RadiusKm:  session.RadiusKm,
		Dishes:    feedItems,
	}

	// Construir metadatos de paginación
	meta := dto.NewPaginationMeta(page, perPage, totalItems)

	// Construir respuesta paginada
	paginatedData := dto.PaginatedData{
		Data: feedData,
		Meta: meta,
	}

	resp := dto.SuccessResponse(
		dto.WithData(paginatedData),
		dto.WithMessage("Feed de platos obtenido exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// GetActiveSession godoc
// @Summary Obtiene la sesión activa del usuario
// @Description Retorna la sesión de swipe activa actual del usuario autenticado
// @Tags sessions
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.SessionData}
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /sessions/active [get]
func (c *SwipeSessionController) GetActiveSession(ctx *gin.Context) {
	// Obtener el userID del contexto
	userIDStr := ctx.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "ID de usuario inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Obtener sesión activa
	session, err := c.sessionService.GetActiveByUserID(ctx, userID)
	if err != nil {
		resp := dto.ErrorResponse(404, "NOT_FOUND: No tienes una sesión activa", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusNotFound, resp)
		return
	}

	// Construir respuesta exitosa
	data := dto.SessionData{
		ID:        session.ID,
		UserLat:   session.UserLat,
		UserLng:   session.UserLng,
		RadiusKm:  session.RadiusKm,
		Status:    string(session.Status),
		StartedAt: session.StartedAt.Format("2006-01-02T15:04:05Z"),
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Sesión activa encontrada"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// Complete godoc
// @Summary Completa una sesión de swipe
// @Description Marca la sesión como completada
// @Tags sessions
// @Produce json
// @Param id path string true "ID de la sesión (UUID)"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /sessions/{id}/complete [post]
func (c *SwipeSessionController) Complete(ctx *gin.Context) {
	// Obtener el userID del contexto
	userIDStr := ctx.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "ID de usuario inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Obtener el ID de la sesión de la URL
	sessionIDStr := ctx.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de sesión inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// Completar la sesión
	if err := c.sessionService.Complete(ctx, sessionID, userID); err != nil {
		resp := dto.ErrorResponse(500, "INTERNAL_ERROR: Error al completar sesión", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := dto.SuccessResponse(
		dto.WithMessage("Sesión completada exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// SetupRoutes configura las rutas de sesiones con middleware de autenticación
func (c *SwipeSessionController) SetupRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	sessions := router.Group("/sessions")
	sessions.Use(authMiddleware)
	{
		sessions.POST("", middleware.ValidationMiddleware(&dto.CreateSessionRequest{}), c.Create)
		sessions.GET("/active", c.GetActiveSession)
		sessions.POST("/:id/complete", c.Complete)
		sessions.GET("/:id/feed", c.GetFeed)
		sessions.POST("/:id/swipe", middleware.ValidationMiddleware(&dto.SwipeRequest{}), c.Swipe)
		sessions.GET("/:id/finalists", c.GetFinalists)
		sessions.POST("/:id/match", middleware.ValidationMiddleware(&dto.MatchRequest{}), c.Match)
	}
}

// Swipe godoc
// @Summary Registra un swipe en una sesión
// @Description Registra un swipe (like/pass) en un plato dentro de una sesión activa
// @Tags sessions
// @Accept json
// @Produce json
// @Param id path string true "ID de la sesión (UUID)"
// @Param request body dto.SwipeRequest true "Datos del swipe"
// @Success 200 {object} dto.APIResponse{data=dto.SwipeResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /sessions/{id}/swipe [post]
func (c *SwipeSessionController) Swipe(ctx *gin.Context) {
	userIDStr := ctx.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "ID de usuario inválido")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	sessionIDStr := ctx.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de sesión inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	req := middleware.GetValidatedRequest(ctx).(*dto.SwipeRequest)

	dishID, err := uuid.Parse(req.DishID)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_DISH_ID: ID de plato inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	var direction domain.SwipeDirection
	if req.Direction == "right" {
		direction = domain.SwipeDirectionRight
	} else {
		direction = domain.SwipeDirectionLeft
	}

	result, err := c.sessionService.Swipe(ctx, sessionID, userID, dishID, direction)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMaxLikesReached):
			resp := dto.ErrorResponse(403, "MAX_LIKES_REACHED: "+err.Error())
			ctx.JSON(http.StatusForbidden, resp)
		case errors.Is(err, domain.ErrDishAlreadySwiped):
			resp := dto.ErrorResponse(400, "DISH_ALREADY_SWIPED: "+err.Error())
			ctx.JSON(http.StatusBadRequest, resp)
		default:
			resp := dto.ErrorResponse(400, "SWIPE_ERROR: "+err.Error())
			ctx.JSON(http.StatusBadRequest, resp)
		}
		return
	}

	data := dto.SwipeResponse{
		SwipeID:     result.SwipeID.String(),
		LikesCount:  result.LikesCount,
		FeedBlocked: result.FeedBlocked,
	}

	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Swipe registrado exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// GetFinalists godoc
// @Summary Obtiene los finalistas de una sesión
// @Description Retorna los 3 platos con like de una sesión
// @Tags sessions
// @Produce json
// @Param id path string true "ID de la sesión (UUID)"
// @Success 200 {object} dto.APIResponse{data=dto.FinalistsResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /sessions/{id}/finalists [get]
func (c *SwipeSessionController) GetFinalists(ctx *gin.Context) {
	userIDStr := ctx.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "ID de usuario inválido")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	sessionIDStr := ctx.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de sesión inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	result, err := c.sessionService.GetFinalists(ctx, sessionID, userID)
	if err != nil {
		resp := dto.ErrorResponse(400, "FINALISTS_ERROR: "+err.Error())
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	var finalists []dto.FinalistDish
	for _, f := range result.Finalists {
		finalists = append(finalists, dto.FinalistDish{
			ID:          f.ID,
			Name:        f.Name,
			Description: f.Description,
			Price:       f.Price,
			PhotoURL:    f.PhotoURL,
			Company: dto.FinalistCompany{
				ID:      f.Company.ID,
				Name:    f.Company.Name,
				Address: f.Company.Address,
				Lat:     f.Company.Lat,
				Lng:     f.Company.Lng,
				LogoURL: f.Company.LogoURL,
			},
			SwipedAt: f.SwipedAt,
		})
	}

	data := dto.FinalistsResponse{
		SessionID: result.SessionID.String(),
		Finalists: finalists,
		Count:     result.Count,
	}

	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Finalistas obtenidos exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// Match godoc
// @Summary Crea un match con un plato finalista
// @Description Selecciona un plato de los finalistas y completa la sesión
// @Tags sessions
// @Accept json
// @Produce json
// @Param id path string true "ID de la sesión (UUID)"
// @Param request body dto.MatchRequest true "Datos del match"
// @Success 200 {object} dto.APIResponse{data=dto.MatchResponse}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 403 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /sessions/{id}/match [post]
func (c *SwipeSessionController) Match(ctx *gin.Context) {
	userIDStr := ctx.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "ID de usuario inválido")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	sessionIDStr := ctx.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de sesión inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	req := middleware.GetValidatedRequest(ctx).(*dto.MatchRequest)

	dishID, err := uuid.Parse(req.DishID)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_DISH_ID: ID de plato inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	result, err := c.sessionService.Match(ctx, sessionID, userID, dishID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrSessionAlreadyCompleted):
			resp := dto.ErrorResponse(400, "SESSION_ALREADY_COMPLETED: "+err.Error())
			ctx.JSON(http.StatusBadRequest, resp)
		case errors.Is(err, domain.ErrDishNotInFinalists):
			resp := dto.ErrorResponse(400, "DISH_NOT_IN_FINALISTS: "+err.Error())
			ctx.JSON(http.StatusBadRequest, resp)
		case errors.Is(err, domain.ErrUnauthorizedSession):
			resp := dto.ErrorResponse(401, "UNAUTHORIZED: "+err.Error())
			ctx.JSON(http.StatusUnauthorized, resp)
		default:
			resp := dto.ErrorResponse(400, "MATCH_ERROR: "+err.Error())
			ctx.JSON(http.StatusBadRequest, resp)
		}
		return
	}

	data := dto.MatchResponse{
		Dish: dto.MatchDish{
			ID:       result.Dish.ID.String(),
			Name:     result.Dish.Name,
			Price:    result.Dish.Price,
			PhotoURL: result.Dish.PhotoURL,
		},
		Company: dto.MatchCompany{
			Name:    result.Company.Name,
			Address: result.Company.Address,
			Phone:   result.Company.Phone,
			Lat:     result.Company.Lat,
			Lng:     result.Company.Lng,
		},
		MatchedAt: result.MatchedAt.Format("2006-01-02T15:04:05Z"),
	}

	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Match creado exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}
