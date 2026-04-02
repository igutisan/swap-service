package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"swap/iguti/swap-service/internal/domain"
	"swap/iguti/swap-service/internal/dto"
	"swap/iguti/swap-service/internal/middleware"
)

// DishController maneja los endpoints de platos
type DishController struct {
	dishService domain.DishService
}

// NewDishController crea una nueva instancia de DishController
func NewDishController(dishService domain.DishService) *DishController {
	return &DishController{
		dishService: dishService,
	}
}

// Create godoc
// @Summary Crea un nuevo plato
// @Description Crea un plato para la compañía autenticada
// @Tags dishes
// @Accept json
// @Produce json
// @Param request body dto.CreateDishRequest true "Datos del plato"
// @Success 201 {object} dto.APIResponse{data=dto.DishData}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /dishes [post]
func (c *DishController) Create(ctx *gin.Context) {
	// Obtener el companyID del contexto (inyectado por el middleware JWT de empresa)
	companyIDStr := ctx.GetString("companyID")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "UNAUTHORIZED: ID de compañía inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Obtener request validado del contexto
	req := middleware.GetValidatedRequest(ctx).(*dto.CreateDishRequest)

	// Llamar al servicio
	dish, err := c.dishService.Create(ctx, companyID, req.Name, req.Description, req.Price, req.PhotoURL)
	if err != nil {
		resp := dto.ErrorResponse(500, "INTERNAL_ERROR", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Construir respuesta exitosa
	data := dto.DishData{
		ID:          dish.ID,
		CompanyID:   dish.CompanyID,
		Name:        dish.Name,
		Description: dish.Description,
		Price:       dish.Price,
		PhotoURL:    dish.PhotoURL,
		IsActive:    dish.IsActive,
		CreatedAt:   dish.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Plato creado exitosamente"),
	)
	ctx.JSON(http.StatusCreated, resp)
}

// GetByID godoc
// @Summary Obtiene un plato por ID
// @Description Obtiene los detalles de un plato específico
// @Tags dishes
// @Produce json
// @Param id path string true "ID del plato (UUID)"
// @Success 200 {object} dto.APIResponse{data=dto.DishData}
// @Failure 400 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /dishes/{id} [get]
func (c *DishController) GetByID(ctx *gin.Context) {
	// Obtener el ID del plato de la URL
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de plato inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// Llamar al servicio
	dish, err := c.dishService.GetByID(ctx, id)
	if err != nil {
		resp := dto.ErrorResponse(404, "NOT_FOUND: Plato no encontrado", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusNotFound, resp)
		return
	}

	// Construir respuesta exitosa
	data := dto.DishData{
		ID:          dish.ID,
		CompanyID:   dish.CompanyID,
		Name:        dish.Name,
		Description: dish.Description,
		Price:       dish.Price,
		PhotoURL:    dish.PhotoURL,
		IsActive:    dish.IsActive,
		CreatedAt:   dish.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Plato encontrado"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// GetByCompanyID godoc
// @Summary Obtiene todos los platos de una compañía
// @Description Obtiene la lista de todos los platos de la compañía autenticada
// @Tags dishes
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.DishListData}
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /dishes [get]
func (c *DishController) GetByCompanyID(ctx *gin.Context) {
	// Obtener el companyID del contexto
	companyIDStr := ctx.GetString("companyID")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "UNAUTHORIZED: ID de compañía inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Llamar al servicio
	dishes, err := c.dishService.GetByCompanyID(ctx, companyID)
	if err != nil {
		resp := dto.ErrorResponse(500, "INTERNAL_ERROR", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Convertir a DTOs
	var dishesData []dto.DishData
	for _, dish := range dishes {
		dishesData = append(dishesData, dto.DishData{
			ID:          dish.ID,
			CompanyID:   dish.CompanyID,
			Name:        dish.Name,
			Description: dish.Description,
			Price:       dish.Price,
			PhotoURL:    dish.PhotoURL,
			IsActive:    dish.IsActive,
			CreatedAt:   dish.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	data := dto.DishListData{
		Dishes: dishesData,
		Count:  len(dishesData),
	}

	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Platos obtenidos exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// GetActiveByCompanyID godoc
// @Summary Obtiene solo los platos activos de una compañía
// @Description Obtiene la lista de platos activos de la compañía autenticada
// @Tags dishes
// @Produce json
// @Success 200 {object} dto.APIResponse{data=dto.DishListData}
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /dishes/active [get]
func (c *DishController) GetActiveByCompanyID(ctx *gin.Context) {
	// Obtener el companyID del contexto
	companyIDStr := ctx.GetString("companyID")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "UNAUTHORIZED: ID de compañía inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Llamar al servicio
	dishes, err := c.dishService.GetActiveByCompanyID(ctx, companyID)
	if err != nil {
		resp := dto.ErrorResponse(500, "INTERNAL_ERROR", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Convertir a DTOs
	var dishesData []dto.DishData
	for _, dish := range dishes {
		dishesData = append(dishesData, dto.DishData{
			ID:          dish.ID,
			CompanyID:   dish.CompanyID,
			Name:        dish.Name,
			Description: dish.Description,
			Price:       dish.Price,
			PhotoURL:    dish.PhotoURL,
			IsActive:    dish.IsActive,
			CreatedAt:   dish.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	data := dto.DishListData{
		Dishes: dishesData,
		Count:  len(dishesData),
	}

	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Platos activos obtenidos exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// Update godoc
// @Summary Actualiza un plato
// @Description Actualiza los datos de un plato existente
// @Tags dishes
// @Accept json
// @Produce json
// @Param id path string true "ID del plato (UUID)"
// @Param request body dto.UpdateDishRequest true "Datos a actualizar"
// @Success 200 {object} dto.APIResponse{data=dto.DishData}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /dishes/{id} [put]
func (c *DishController) Update(ctx *gin.Context) {
	// Obtener el ID del plato de la URL
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de plato inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// Obtener request validado del contexto
	req := middleware.GetValidatedRequest(ctx).(*dto.UpdateDishRequest)

	// Llamar al servicio
	dish, err := c.dishService.Update(ctx, id, req.Name, req.Description, req.Price, req.PhotoURL)
	if err != nil {
		resp := dto.ErrorResponse(500, "INTERNAL_ERROR", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Construir respuesta exitosa
	data := dto.DishData{
		ID:          dish.ID,
		CompanyID:   dish.CompanyID,
		Name:        dish.Name,
		Description: dish.Description,
		Price:       dish.Price,
		PhotoURL:    dish.PhotoURL,
		IsActive:    dish.IsActive,
		CreatedAt:   dish.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Plato actualizado exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// Delete godoc
// @Summary Elimina un plato
// @Description Elimina permanentemente un plato de la compañía
// @Tags dishes
// @Produce json
// @Param id path string true "ID del plato (UUID)"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /dishes/{id} [delete]
func (c *DishController) Delete(ctx *gin.Context) {
	// Obtener el ID del plato de la URL
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de plato inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// Obtener el companyID del contexto
	companyIDStr := ctx.GetString("companyID")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "UNAUTHORIZED: ID de compañía inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Llamar al servicio
	if err := c.dishService.Delete(ctx, id, companyID); err != nil {
		resp := dto.ErrorResponse(500, "INTERNAL_ERROR", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	resp := dto.SuccessResponse(
		dto.WithMessage("Plato eliminado exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// ToggleActive godoc
// @Summary Activa o desactiva un plato
// @Description Cambia el estado activo/inactivo de un plato sin eliminarlo
// @Tags dishes
// @Produce json
// @Param id path string true "ID del plato (UUID)"
// @Success 200 {object} dto.APIResponse{data=dto.ToggleDishStatusData}
// @Failure 400 {object} dto.APIResponse
// @Failure 401 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /dishes/{id}/toggle [patch]
func (c *DishController) ToggleActive(ctx *gin.Context) {
	// Obtener el ID del plato de la URL
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		resp := dto.ErrorResponse(400, "INVALID_ID: ID de plato inválido")
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}

	// Obtener el companyID del contexto
	companyIDStr := ctx.GetString("companyID")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		resp := dto.ErrorResponse(401, "UNAUTHORIZED: ID de compañía inválido en token")
		ctx.JSON(http.StatusUnauthorized, resp)
		return
	}

	// Llamar al servicio
	dish, err := c.dishService.ToggleActive(ctx, id, companyID)
	if err != nil {
		resp := dto.ErrorResponse(500, "INTERNAL_ERROR", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Determinar el estado en texto
	status := "desactivado"
	if dish.IsActive {
		status = "activado"
	}

	data := dto.ToggleDishStatusData{
		ID:       dish.ID,
		IsActive: dish.IsActive,
		Status:   status,
	}

	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("Plato "+status+" exitosamente"),
	)
	ctx.JSON(http.StatusOK, resp)
}

// SetupRoutes configura las rutas de platos con middleware de validación y autenticación
func (c *DishController) SetupRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	dishes := router.Group("/dishes")
	dishes.Use(authMiddleware) // Aplicar middleware de autenticación JWT
	{
		// Rutas de colección
		dishes.GET("", c.GetByCompanyID)
		dishes.GET("/active", c.GetActiveByCompanyID)
		dishes.POST("", middleware.ValidationMiddleware(&dto.CreateDishRequest{}), c.Create)

		// Rutas de recurso individual
		dishes.GET("/:id", c.GetByID)
		dishes.PUT("/:id", middleware.ValidationMiddleware(&dto.UpdateDishRequest{}), c.Update)
		dishes.DELETE("/:id", c.Delete)
		dishes.PATCH("/:id/toggle", c.ToggleActive)
	}
}
