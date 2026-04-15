package service

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"swap/iguti/swap-service/internal/config"
	"swap/iguti/swap-service/internal/dto"
)

// UploadService define las operaciones para subida de archivos
type UploadService interface {
	GeneratePresignedURL(ctx context.Context, folder, fileName, fileType string) (*dto.PresignedURLData, error)
}

// CloudinaryService implementa UploadService usando Cloudinary
type cloudinaryService struct {
	config *config.CloudinaryConfig
	cld    *cloudinary.Cloudinary
}

// NewCloudinaryService crea una nueva instancia del servicio de Cloudinary
func NewCloudinaryService(cfg *config.CloudinaryConfig) (UploadService, error) {
	if !cfg.IsConfigured() {
		return nil, fmt.Errorf("cloudinary no está configurado correctamente")
	}

	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("error al crear cliente de cloudinary: %w", err)
	}

	return &cloudinaryService{
		config: cfg,
		cld:    cld,
	}, nil
}

// GeneratePresignedURL genera una URL y parámetros firmados para subir directamente a Cloudinary
func (s *cloudinaryService) GeneratePresignedURL(ctx context.Context, folder, fileName, fileType string) (*dto.PresignedURLData, error) {
	// Generar timestamp y public_id
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	publicID := generatePublicID(fileName)

	// Si se especifica una carpeta, agregarla al public_id
	if folder != "" {
		publicID = folder + "/" + publicID
	}

	// Construir string para firmar
	signatureString := fmt.Sprintf("public_id=%s&timestamp=%s%s", publicID, timestamp, s.config.APISecret)
	signature := sha1Hash(signatureString)

	// Construir URL de upload
	uploadURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", s.config.CloudName)

	// Preparar campos del formulario
	fields := map[string]string{
		"api_key":   s.config.APIKey,
		"timestamp": timestamp,
		"public_id": publicID,
		"signature": signature,
	}

	return &dto.PresignedURLData{
		URL:       uploadURL,
		Fields:    fields,
		PublicID:  publicID,
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}, nil
}

// sha1Hash genera un hash SHA1 del string proporcionado
func sha1Hash(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// generatePublicID genera un public_id único basado en el nombre del archivo
func generatePublicID(fileName string) string {
	// Agregar timestamp para hacerlo único
	timestamp := time.Now().Unix()
	return fmt.Sprintf("dish_%d_%s", timestamp, sanitizeFileName(fileName))
}

// sanitizeFileName limpia el nombre del archivo para uso seguro
func sanitizeFileName(fileName string) string {
	// Reemplazar caracteres no permitidos
	result := ""
	for _, char := range fileName {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			result += string(char)
		} else {
			result += "_"
		}
	}
	return result
}
