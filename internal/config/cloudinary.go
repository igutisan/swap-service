package config

import (
	"os"
)

// CloudinaryConfig contiene la configuración para Cloudinary
type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
}

// NewCloudinaryConfig crea una configuración de Cloudinary desde variables de entorno
func NewCloudinaryConfig() *CloudinaryConfig {
	return &CloudinaryConfig{
		CloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		APIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		APISecret: getEnv("CLOUDINARY_API_SECRET", ""),
	}
}

// IsConfigured verifica si la configuración está completa
func (c *CloudinaryConfig) IsConfigured() bool {
	return c.CloudName != "" && c.APIKey != "" && c.APISecret != ""
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
