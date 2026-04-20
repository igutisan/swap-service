package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"swap/iguti/swap-service/internal/domain"
)

const (
	mapboxBaseURL = "https://api.mapbox.com/search/geocode/v6/forward"
	httpTimeout   = 10 * time.Second
)

type GeocodingResult struct {
	Lat     float64
	Lng     float64
	Name    string
	City    string
	Country string
}

type mapboxFeature struct {
	Geometry struct {
		Coordinates [2]float64 `json:"coordinates"`
	} `json:"geometry"`
	Properties mapboxProperties `json:"properties"`
}

type mapboxProperties struct {
	Name             string `json:"name"`
	PlaceFormatted   string `json:"place_formatted"`
	Coordinates     mapboxCoords `json:"coordinates"`
	Context         mapboxContext `json:"context"`
}

type mapboxCoords struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type mapboxContext struct {
	Country mapboxContextItem `json:"country"`
	Region  mapboxContextItem `json:"region"`
	Place   mapboxContextItem `json:"place"`
	Postcode mapboxContextItem `json:"postcode"`
}

type mapboxContextItem struct {
	Name         string `json:"name"`
	CountryCode  string `json:"country_code"`
	RegionCode  string `json:"region_code"`
}

type mapboxResponse struct {
	Type     string          `json:"type"`
	Features []mapboxFeature `json:"features"`
}

type GeocodingService interface {
	Geocode(ctx context.Context, address string) (*GeocodingResult, error)
	ValidateCoordinates(lat, lng float64) error
}

type geocodingService struct {
	httpClient *http.Client
	apiKey     string
}

func NewGeocodingService() GeocodingService {
	apiKey := os.Getenv("GEO_API")
	if apiKey == "" {
		fmt.Println("[WARNING] GEO_API no configurada, el geocoding no funcionará")
	}
	return &geocodingService{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		apiKey: apiKey,
	}
}

func (s *geocodingService) Geocode(ctx context.Context, address string) (*GeocodingResult, error) {
	if address == "" {
		return nil, domain.ErrGeocodingEmptyAddress
	}

	if s.apiKey == "" {
		return nil, fmt.Errorf("GEO_API no configurada en el entorno")
	}

	encodedAddress := url.QueryEscape(strings.TrimSpace(address))
	requestURL := fmt.Sprintf(
		"%s?q=%s&access_token=%s&country=CO&limit=1&language=es&types=address,place,street",
		mapboxBaseURL, encodedAddress, s.apiKey,
	)

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

	var mapboxResp mapboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&mapboxResp); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta de geocoding: %w", err)
	}

	if len(mapboxResp.Features) == 0 {
		return nil, domain.ErrGeocodingNotFound
	}

	feature := mapboxResp.Features[0]

	var lng, lat float64
	if feature.Properties.Coordinates.Longitude != 0 {
		lng = feature.Properties.Coordinates.Longitude
		lat = feature.Properties.Coordinates.Latitude
	} else if len(feature.Geometry.Coordinates) == 2 {
		lng = feature.Geometry.Coordinates[0]
		lat = feature.Geometry.Coordinates[1]
	} else {
		return nil, domain.ErrGeocodingNotFound
	}

	if err := s.ValidateCoordinates(lat, lng); err != nil {
		return nil, err
	}

	city := feature.Properties.Context.Place.Name
	if city == "" {
		city = "Colombia"
	}

	country := feature.Properties.Context.Country.Name
	if country == "" {
		country = feature.Properties.Context.Country.CountryCode
	}

	return &GeocodingResult{
		Lat:     lat,
		Lng:     lng,
		Name:    feature.Properties.Name,
		City:    city,
		Country: country,
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
