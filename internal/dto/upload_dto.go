package dto

// GeneratePresignedURLRequest representa la petición para generar una URL presigned
type GeneratePresignedURLRequest struct {
	Folder   string `json:"folder" validate:"omitempty,max=100"`
	FileName string `json:"file_name" validate:"required,max=255"`
	FileType string `json:"file_type" validate:"required,oneof=image/jpeg image/png image/webp"`
}

// PresignedURLData representa los datos de la URL presigned generada
type PresignedURLData struct {
	URL       string            `json:"url"`
	Fields    map[string]string `json:"fields"`
	PublicID  string            `json:"public_id"`
	ExpiresAt int64             `json:"expires_at"`
}

// UploadImageCallbackRequest representa la petición de callback después de subir la imagen
type UploadImageCallbackRequest struct {
	PublicID  string `json:"public_id" validate:"required"`
	SecureURL string `json:"secure_url" validate:"required,url"`
	DishID    string `json:"dish_id,omitempty" validate:"omitempty,uuid"`
}

// ImageUploadedData representa los datos después de confirmar la subida
type ImageUploadedData struct {
	PublicID  string `json:"public_id"`
	SecureURL string `json:"secure_url"`
	DishID    string `json:"dish_id,omitempty"`
	Message   string `json:"message"`
}
