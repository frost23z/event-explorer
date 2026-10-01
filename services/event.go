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
	events, _, err := s.lookup(ctx, params)

	return events, err
}

// lookup is Events plus whether the list came from the cache.
func (s *EventService) lookup(ctx context.Context, params models.EventSearchParams) ([]models.Event, bool, error) {
	key := cacheKey(params)

	if events, ok := s.cache.get(key); ok {
		return events, true, nil
	}

	events, err := s.client.Events(ctx, params)
	if err != nil {
		return nil, false, err
	}

	s.cache.set(key, events)

	return events, false, nil
}

// ClearCache drops every cached list, so the next search asks Ticketmaster
// again. It returns how many entries were removed.
func (s *EventService) ClearCache() int {
	return s.cache.clear()
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
		events, cached, err := s.lookup(ctx, params)
		ch <- models.CategoryResult{Events: events, Cached: cached, Err: err}
	}()

	return ch
}
