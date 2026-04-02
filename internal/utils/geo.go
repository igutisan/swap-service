package utils

import "math"

type BoundingBox struct {
	MinLat float64
	MaxLat float64
	MinLng float64
	MaxLng float64
}

func CalculateBoundingBox(lat, lng float64, radiusKm float64) BoundingBox {
	latDelta := radiusKm / 111.0
	lngDelta := radiusKm / (111.0 * math.Cos(lat*math.Pi/180))

	return BoundingBox{
		MinLat: lat - latDelta,
		MaxLat: lat + latDelta,
		MinLng: lng - lngDelta,
		MaxLng: lng + lngDelta,
	}
}
