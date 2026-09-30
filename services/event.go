package services

import (
	"context"

	"event-explorer/models"
	"event-explorer/services/ticketmasterapi"
)

type EventService struct {
	client *ticketmasterapi.Client
	cache  *eventCache
}

func NewEventService(client *ticketmasterapi.Client) *EventService {
	return &EventService{client: client, cache: newEventCache()}
}

// Events returns the cached list when there is a fresh one. Only successful
// lists are cached, so a failure is retried on the next request.
func (s *EventService) Events(ctx context.Context, params models.EventSearchParams) ([]models.Event, error) {
	key := cacheKey(params)

	if events, ok := s.cache.get(key); ok {
		return events, nil
	}

	events, err := s.client.Events(ctx, params)
	if err != nil {
		return nil, err
	}

	s.cache.set(key, events)

	return events, nil
}

func (s *EventService) Event(ctx context.Context, eventID string) (models.Event, error) {
	return s.client.Event(ctx, eventID)
}

func (s *EventService) MusicAndSports(ctx context.Context, city, countryCode string) (music, sports models.CategoryResult) {
	musicCh := s.eventsAsync(ctx, models.EventSearchParams{
		City: city, CountryCode: countryCode, Category: models.CategoryMusic,
	})
	sportsCh := s.eventsAsync(ctx, models.EventSearchParams{
		City: city, CountryCode: countryCode, Category: models.CategorySports,
	})

	return <-musicCh, <-sportsCh
}

func (s *EventService) eventsAsync(ctx context.Context, params models.EventSearchParams) <-chan models.CategoryResult {
	ch := make(chan models.CategoryResult, 1)

	go func() {
		events, err := s.Events(ctx, params)
		ch <- models.CategoryResult{Events: events, Err: err}
	}()

	return ch
}
