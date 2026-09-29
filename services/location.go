package services

import (
	"context"

	"event-explorer/models"
	"event-explorer/services/googleplacesapi"
)

type LocationService struct {
	client *googleplacesapi.Client
}

func NewLocationService(client *googleplacesapi.Client) *LocationService {
	return &LocationService{client: client}
}

func (s *LocationService) Autocomplete(ctx context.Context, params models.AutocompleteParams) (models.AutocompleteResponse, error) {
	return s.client.Autocomplete(ctx, params)
}

func (s *LocationService) PlaceDetails(ctx context.Context, params models.PlaceDetailsParams) (models.City, error) {
	return s.client.PlaceDetails(ctx, params)
}
