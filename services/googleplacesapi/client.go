package googleplacesapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"event-explorer/models"

	"github.com/beego/beego/v2/core/logs"
)

const (
	defaultBaseURL   = "https://places.googleapis.com"
	defaultTimeout   = 5 * time.Second
	autocompletePath = "/v1/places:autocomplete"
	placesPath       = "/v1/places/"
	userAgent        = "event-explorer/1.0"
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

func (c *Client) Autocomplete(ctx context.Context, params models.AutocompleteParams) (models.AutocompleteResponse, error) {
	//nolint:gosec // Google Places requires the session token in the request body.
	body, err := json.Marshal(autocompleteRequest{
		Input:                params.Input,
		IncludedPrimaryTypes: []string{"(cities)"},
		SessionToken:         params.SessionToken,
		LanguageCode:         "en",
	})
	if err != nil {
		return models.AutocompleteResponse{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.BaseURL+autocompletePath,
		bytes.NewReader(body),
	)
	if err != nil {
		return models.AutocompleteResponse{}, err
	}

	res, err := c.send(req)
	if err != nil {
		return models.AutocompleteResponse{}, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logs.Warn("failed to close response body: %v", err)
		}
	}()

	var result autocompleteResponse
	if err := readJSON(res, &result); err != nil {
		return models.AutocompleteResponse{}, err
	}

	suggestions := make([]models.Suggestion, 0, MaxAutocompleteSuggestions)
	for _, s := range result.Suggestions {
		if s.PlacePrediction == nil || s.PlacePrediction.PlaceID == "" {
			continue
		}

		suggestions = append(suggestions, models.Suggestion{
			PlaceID: s.PlacePrediction.PlaceID,
			Text:    s.PlacePrediction.Text.Text,
		})

		if len(suggestions) == MaxAutocompleteSuggestions {
			break
		}
	}

	return models.AutocompleteResponse{Suggestions: suggestions}, nil
}

func (c *Client) PlaceDetails(ctx context.Context, params models.PlaceDetailsParams) (models.City, error) {
	query := url.Values{"sessionToken": {params.SessionToken}, "languageCode": {"en"}}
	endpoint := c.BaseURL + placesPath + url.PathEscape(params.PlaceID) + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return models.City{}, err
	}

	req.Header.Set("X-Goog-FieldMask", "addressComponents")

	res, err := c.send(req)
	if err != nil {
		return models.City{}, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			logs.Warn("failed to close response body: %v", err)
		}
	}()

	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusBadRequest {
		return models.City{}, cityNotFound(
			fmt.Errorf("google places api returned %s", res.Status),
		)
	}

	var result placeResponse
	if err := readJSON(res, &result); err != nil {
		return models.City{}, err
	}

	var city, postalTown, level3, country string

	for _, comp := range result.AddressComponents {
		for _, t := range comp.Types {
			switch t {
			case "locality":
				city = comp.LongText
			case "postal_town":
				postalTown = comp.LongText
			case "administrative_area_level_3":
				level3 = comp.LongText
			case "country":
				country = strings.ToUpper(comp.ShortText)
			}
		}
	}

	for _, fallback := range []string{postalTown, level3} {
		if city == "" {
			city = fallback
		}
	}

	if city == "" || len(country) != 2 {
		return models.City{}, cityNotFound(errors.New("place has no city or country code"))
	}

	return models.City{City: city, CountryCode: country}, nil
}

func (c *Client) send(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", c.APIKey)
	req.Header.Set("User-Agent", userAgent)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		if errors.Is(req.Context().Err(), context.Canceled) {
			return nil, req.Context().Err()
		}

		return nil, providerError(err)
	}

	return res, nil
}

func readJSON(res *http.Response, out any) error {
	if res.StatusCode != http.StatusOK {
		return providerError(fmt.Errorf("google places api returned %s", res.Status))
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return providerError(fmt.Errorf("decode Google Places response: %w", err))
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

func cityNotFound(err error) models.APIError {
	return models.APIError{
		StatusCode: http.StatusNotFound,
		Message:    MsgCityNotFound,
		Err:        err,
	}
}
