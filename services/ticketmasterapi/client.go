package ticketmasterapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"event-explorer/models"

	"github.com/beego/beego/v2/core/logs"
)

const (
	defaultBaseURL = "https://app.ticketmaster.com/discovery/v2"
	defaultTimeout = 5 * time.Second
	userAgent      = "event-explorer/1.0"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL:    defaultBaseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: defaultTimeout},
	}
}

func (c *Client) Events(ctx context.Context, params models.EventSearchParams) ([]models.Event, error) {
	query := url.Values{
		"city":               {params.City},
		"countryCode":        {params.CountryCode},
		"classificationName": {params.Category},
		"size":               {strconv.Itoa(EventsPageSize)},
	}

	res, err := c.get(ctx, "/events.json", query)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logs.Warn("failed to close response body: %v", err)
		}
	}()

	var result eventsResponse
	if err := readJSON(res, &result); err != nil {
		return nil, err
	}

	events := make([]models.Event, 0, EventsPageSize)

	for _, p := range result.Embedded.Events {
		if p.ID == "" || p.Name == "" {
			continue
		}

		events = append(events, p.toEvent())

		if len(events) == EventsPageSize {
			break
		}
	}

	return events, nil
}

// Event returns one event, including its ticket URL.
func (c *Client) Event(ctx context.Context, eventID string) (models.Event, error) {
	res, err := c.get(ctx, "/events/"+url.PathEscape(eventID)+".json", nil)
	if err != nil {
		return models.Event{}, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logs.Warn("failed to close response body: %v", err)
		}
	}()

	// Ticketmaster answers 404 for an unknown event and 400 for a malformed ID.
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusBadRequest {
		return models.Event{}, eventNotFound(fmt.Errorf("ticketmaster api returned %s", res.Status))
	}

	var payload eventPayload
	if err := readJSON(res, &payload); err != nil {
		return models.Event{}, err
	}

	if payload.ID == "" || payload.Name == "" {
		return models.Event{}, eventNotFound(errors.New("event has no id or name"))
	}

	return payload.toEvent(), nil
}

func (p eventPayload) toEvent() models.Event {
	event := models.Event{
		ID:          p.ID,
		Name:        p.Name,
		ImageURL:    p.pickImage(),
		Date:        p.Dates.Start.LocalDate,
		Time:        p.Dates.Start.LocalTime,
		Description: firstNonEmpty(p.Info, p.PleaseNote, p.Description),
		TicketURL:   p.URL,
	}

	if len(p.Embedded.Venues) > 0 {
		venue := p.Embedded.Venues[0]
		event.Venue = venue.Name
		event.Address = venue.Address.Line1
		event.City = venue.City.Name
	}

	return event
}

func (p eventPayload) pickImage() string {
	var best, widest, first string

	bestWidth, widestWidth := 0, 0

	for _, img := range p.Images {
		if img.URL == "" {
			continue
		}

		if first == "" {
			first = img.URL
		}

		if img.Ratio != imageRatio {
			continue
		}

		if img.Width >= imageMinWidth && (best == "" || img.Width < bestWidth) {
			best, bestWidth = img.URL, img.Width
		}

		if widest == "" || img.Width > widestWidth {
			widest, widestWidth = img.URL, img.Width
		}
	}

	return firstNonEmpty(best, widest, first)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}

func (c *Client) get(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	if query == nil {
		query = url.Values{}
	}

	query.Set("apikey", c.APIKey)

	// Ticketmaster only accepts the key as a query parameter, so any error that
	// embeds the request URL must be stripped before it can reach a log.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path+"?"+query.Encode(), nil)
	if err != nil {
		return nil, errors.New("build ticketmaster request")
	}

	req.Header.Set("User-Agent", userAgent)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, ctx.Err()
		}

		if urlErr, ok := errors.AsType[*url.Error](err); ok {
			err = urlErr.Err
		}

		return nil, providerError(err)
	}

	return res, nil
}

func readJSON(res *http.Response, out any) error {
	if res.StatusCode != http.StatusOK {
		return providerError(fmt.Errorf("ticketmaster api returned %s", res.Status))
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return providerError(fmt.Errorf("decode Ticketmaster response: %w", err))
	}

	return nil
}

func providerError(err error) models.APIError {
	return models.APIError{
		StatusCode: http.StatusBadGateway,
		Message:    MsgProviderFailed,
		Err:        err,
	}
}

func eventNotFound(err error) models.APIError {
	return models.APIError{
		StatusCode: http.StatusNotFound,
		Message:    MsgEventNotFound,
		Err:        err,
	}
}
