package utils

import (
	"math"
)

const earthRadiusKm = 6371 // Radio de la Tierra en kilómetros

// CalculateDistance calcula la distancia entre dos puntos usando la fórmula de Haversine
func CalculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Convertir a radianes
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	// Fórmula de Haversine
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadiusKm * c
	return distance
}

// IsWithinRadius verifica si un punto está dentro de un radio específico
func IsWithinRadius(userLat, userLng, targetLat, targetLng float64, radiusKm int) bool {
	distance := CalculateDistance(userLat, userLng, targetLat, targetLng)
	return distance <= float64(radiusKm)
}

// RoundToDecimals redondea un float64 a n decimales
func RoundToDecimals(value float64, decimals int) float64 {
	p := math.Pow(10, float64(decimals))
	return math.Round(value*p) / p
}
