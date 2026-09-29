package services

import (
	"context"

	"event-explorer/models"
	"event-explorer/services/ticketmasterapi"
)

type EventService struct {
	client *ticketmasterapi.Client
}

func NewEventService(client *ticketmasterapi.Client) *EventService {
	return &EventService{client: client}
}

func (s *EventService) Events(ctx context.Context, params models.EventSearchParams) ([]models.Event, error) {
	return s.client.Events(ctx, params)
}
