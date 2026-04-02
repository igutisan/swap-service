package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"swap/iguti/swap-service/internal/dto"
	"swap/iguti/swap-service/internal/middleware"
	"swap/iguti/swap-service/internal/service"
)

// UploadController maneja los endpoints para subida de archivos
type UploadController struct {
	uploadService service.UploadService
}

// NewUploadController crea una nueva instancia de UploadController
func NewUploadController(uploadService service.UploadService) *UploadController {
	return &UploadController{
		uploadService: uploadService,
	}
}

// GeneratePresignedURL godoc
// @Summary Genera una URL presigned para subir imagen a Cloudinary
// @Description Genera una URL firmada y parámetros para subir directamente a Cloudinary sin pasar por el servidor
// @Tags upload
// @Accept json
// @Produce json
// @Param request body dto.GeneratePresignedURLRequest true "Datos de la imagen a subir"
// @Success 200 {object} dto.APIResponse{data=dto.PresignedURLData}
// @Failure 400 {object} dto.APIResponse
// @Failure 500 {object} dto.APIResponse
// @Router /upload/presigned-url [post]
func (c *UploadController) GeneratePresignedURL(ctx *gin.Context) {
	// Obtener request validado del contexto
	req := middleware.GetValidatedRequest(ctx).(*dto.GeneratePresignedURLRequest)

	// Establecer carpeta por defecto si no se especifica
	folder := req.Folder
	if folder == "" {
		folder = "dishes"
	}

	// Generar URL presigned
	data, err := c.uploadService.GeneratePresignedURL(ctx, folder, req.FileName, req.FileType)
	if err != nil {
		resp := dto.ErrorResponse(500, "UPLOAD_ERROR: Error al generar URL de subida", dto.WithErrors(err.Error()))
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	// Construir respuesta exitosa
	resp := dto.SuccessResponse(
		dto.WithData(data),
		dto.WithMessage("URL presigned generada exitosamente. Use la URL y los campos proporcionados para subir la imagen directamente a Cloudinary via POST multipart/form-data. La URL expira en 15 minutos."),
	)
	ctx.JSON(http.StatusOK, resp)
}

// SetupRoutes configura las rutas de upload con middleware de validación
func (c *UploadController) SetupRoutes(router *gin.RouterGroup) {
	upload := router.Group("/upload")
	{
		// Aplicar middleware de validación antes del handler
		upload.POST("/presigned-url", middleware.ValidationMiddleware(&dto.GeneratePresignedURLRequest{}), c.GeneratePresignedURL)
	}
}
