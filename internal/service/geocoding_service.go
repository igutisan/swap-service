package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"swap/iguti/swap-service/internal/domain"
)

const (
	photonBaseURL = "https://photon.komoot.io/api"
	httpTimeout   = 10 * time.Second
)

type GeocodingResult struct {
	Lat     float64
	Lng     float64
	Name    string
	City    string
	Country string
}

type photonResponse struct {
	Features []photonFeature `json:"features"`
}

type photonFeature struct {
	Geometry   photonGeometry   `json:"geometry"`
	Properties photonProperties `json:"properties"`
}

type photonGeometry struct {
	Coordinates [2]float64 `json:"coordinates"`
}

type photonProperties struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type GeocodingService interface {
	Geocode(ctx context.Context, address string) (*GeocodingResult, error)
	ValidateCoordinates(lat, lng float64) error
}

type geocodingService struct {
	httpClient *http.Client
}

func NewGeocodingService() GeocodingService {
	return &geocodingService{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

func (s *geocodingService) Geocode(ctx context.Context, address string) (*GeocodingResult, error) {
	if address == "" {
		return nil, domain.ErrGeocodingEmptyAddress
	}

	encodedAddress := url.QueryEscape(strings.TrimSpace(address))
	requestURL := fmt.Sprintf("%s?q=%s&limit=1", photonBaseURL, encodedAddress)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error al crear solicitud de geocoding: %w", err)
	}

	req.Header.Set("User-Agent", "SwapService/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al conectar con el servicio de geocoding: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("el servicio de geocoding retornó código de estado: %d", resp.StatusCode)
	}

	var photonResp photonResponse
	if err := json.NewDecoder(resp.Body).Decode(&photonResp); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta de geocoding: %w", err)
	}

	if len(photonResp.Features) == 0 {
		return nil, domain.ErrGeocodingNotFound
	}

	feature := photonResp.Features[0]

	if err := s.ValidateCoordinates(feature.Geometry.Coordinates[1], feature.Geometry.Coordinates[0]); err != nil {
		return nil, err
	}

	return &GeocodingResult{
		Lat:     feature.Geometry.Coordinates[1],
		Lng:     feature.Geometry.Coordinates[0],
		Name:    feature.Properties.Name,
		City:    feature.Properties.City,
		Country: feature.Properties.Country,
	}, nil
}

func (s *geocodingService) ValidateCoordinates(lat, lng float64) error {
	if lat < -90 || lat > 90 {
		return domain.ErrGeocodingInvalidLatitude
	}
	if lng < -180 || lng > 180 {
		return domain.ErrGeocodingInvalidLongitude
	}
	return nil
}
