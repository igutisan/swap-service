package middleware

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"swap/iguti/swap-service/internal/dto"
	"swap/iguti/swap-service/internal/utils"
)

func ValidationMiddleware(requestStruct interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Crear una nueva instancia del tipo de request
		requestType := reflect.TypeOf(requestStruct).Elem()
		requestPtr := reflect.New(requestType)
		request := requestPtr.Interface()

		// Parsear JSON
		if err := c.ShouldBindJSON(request); err != nil {
			resp := dto.ErrorResponse(400, "INVALID_REQUEST: Formato de solicitud inválido", dto.WithErrors(err.Error()))
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}

		// Validar estructura
		validator := utils.NewValidator()
		if validationError := validator.ValidateStruct(request); validationError != nil {
			c.JSON(http.StatusBadRequest, *validationError)
			c.Abort()
			return
		}

		// Guardar el request validado en el contexto para que el handler lo use
		c.Set("validated_request", request)
		c.Next()
	}
}

// GetValidatedRequest obtiene el request validado del contexto
// Uso en handler: req := middleware.GetValidatedRequest(c).(*dto.MyRequest)
func GetValidatedRequest(c *gin.Context) interface{} {
	if req, exists := c.Get("validated_request"); exists {
		return req
	}
	return nil
}
