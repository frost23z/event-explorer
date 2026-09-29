package services

import (
	"context"
	"unicode/utf8"

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
	if utf8.RuneCountInString(params.Input) < models.MinAutocompleteRunes {
		return models.AutocompleteResponse{Suggestions: []models.Suggestion{}}, nil
	}
	return s.client.Autocomplete(ctx, params)
}

func (s *LocationService) PlaceDetails(ctx context.Context, params models.PlaceDetailsParams) (models.City, error) {
	return s.client.PlaceDetails(ctx, params)
}
