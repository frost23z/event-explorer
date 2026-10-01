package googleplacesapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"event-explorer/models"
)

const testAPIKey = "test-key"

// newTestClient returns a client that talks to a fake Google server.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := NewClient(testAPIKey)
	client.BaseURL = server.URL

	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("write response: %v", err)
	}
}

func requireAPIError(t *testing.T, err error, status int, message string) {
	t.Helper()

	apiErr, ok := errors.AsType[models.APIError](err)
	if !ok {
		t.Fatalf("expected models.APIError, got %T: %v", err, err)
	}

	if apiErr.StatusCode != status || apiErr.Message != message {
		t.Fatalf("got %d %q, want %d %q", apiErr.StatusCode, apiErr.Message, status, message)
	}
}

func autocompleteParams() models.AutocompleteParams {
	return models.AutocompleteParams{Input: "toro", SessionToken: "token-1"}
}

func placeParams() models.PlaceDetailsParams {
	return models.PlaceDetailsParams{PlaceID: "place-1", SessionToken: "token-1"}
}

func TestAutocompleteSendsExpectedRequest(t *testing.T) {
	var (
		gotMethod, gotPath, gotKey string
		gotBody                    autocompleteRequest
	)

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotKey = r.Header.Get("X-Goog-Api-Key")

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}

		writeJSON(t, w, `{"suggestions": []}`)
	})

	if _, err := client.Autocomplete(context.Background(), autocompleteParams()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost || gotPath != "/v1/places:autocomplete" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}

	if gotKey != testAPIKey {
		t.Errorf("api key header = %q", gotKey)
	}

	if gotBody.Input != "toro" || gotBody.SessionToken != "token-1" {
		t.Errorf("unexpected body: %+v", gotBody)
	}

	if len(gotBody.IncludedPrimaryTypes) != 1 || gotBody.IncludedPrimaryTypes[0] != "(cities)" {
		t.Errorf("includedPrimaryTypes = %v, want [(cities)]", gotBody.IncludedPrimaryTypes)
	}
}

func TestAutocompleteReturnsSuggestions(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"suggestions": [
			{"placePrediction": {"placeId": "p1", "text": {"text": "Toronto, ON, Canada"}}},
			{"queryPrediction": {"text": {"text": "toronto"}}},
			{"placePrediction": {"placeId": "", "text": {"text": "No ID"}}},
			{"placePrediction": {"placeId": "p2", "text": {"text": "Toronto, OH, USA"}}}
		]}`)
	})

	got, err := client.Autocomplete(context.Background(), autocompleteParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []models.Suggestion{
		{PlaceID: "p1", Text: "Toronto, ON, Canada"},
		{PlaceID: "p2", Text: "Toronto, OH, USA"},
	}

	if len(got.Suggestions) != len(want) {
		t.Fatalf("got %+v, want %+v", got.Suggestions, want)
	}

	for i := range want {
		if got.Suggestions[i] != want[i] {
			t.Errorf("suggestion %d = %+v, want %+v", i, got.Suggestions[i], want[i])
		}
	}
}

func TestAutocompleteCapsSuggestions(t *testing.T) {
	var items []string

	for i := range MaxAutocompleteSuggestions + 3 {
		id := "p" + strconv.Itoa(i)
		items = append(items, `{"placePrediction": {"placeId": "`+id+`", "text": {"text": "City `+id+`"}}}`)
	}

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"suggestions": [`+strings.Join(items, ",")+`]}`)
	})

	got, err := client.Autocomplete(context.Background(), autocompleteParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got.Suggestions) != MaxAutocompleteSuggestions {
		t.Fatalf("got %d suggestions, want %d", len(got.Suggestions), MaxAutocompleteSuggestions)
	}
}

func TestAutocompleteNoSuggestionsIsEmptyList(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{}`)
	})

	got, err := client.Autocomplete(context.Background(), autocompleteParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Suggestions == nil || len(got.Suggestions) != 0 {
		t.Fatalf("want empty non-nil slice so JSON is [], got %#v", got.Suggestions)
	}
}

func TestAutocompleteProviderFailures(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"server error": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		"forbidden key": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		},
		"invalid json": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, `{not json`)
		},
	}

	for name, handler := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, handler)

			_, err := client.Autocomplete(context.Background(), autocompleteParams())
			requireAPIError(t, err, http.StatusBadGateway, MsgProviderFailed)
		})
	}
}

func TestAutocompleteTimeoutIsProviderFailure(t *testing.T) {
	release := make(chan struct{})
	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		<-release
	})
	client.HTTPClient.Timeout = 20 * time.Millisecond

	t.Cleanup(func() { close(release) })

	_, err := client.Autocomplete(context.Background(), autocompleteParams())
	requireAPIError(t, err, http.StatusBadGateway, MsgProviderFailed)
}

func TestAutocompleteCanceledRequestIsNotProviderFailure(t *testing.T) {
	release := make(chan struct{})
	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		<-release
	})

	t.Cleanup(func() { close(release) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Autocomplete(ctx, autocompleteParams())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestPlaceDetailsSendsExpectedRequest(t *testing.T) {
	var gotMethod, gotPath, gotKey, gotMask, gotToken, gotLang string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.EscapedPath()
		gotKey = r.Header.Get("X-Goog-Api-Key")
		gotMask = r.Header.Get("X-Goog-FieldMask")
		gotToken = r.URL.Query().Get("sessionToken")
		gotLang = r.URL.Query().Get("languageCode")

		writeJSON(t, w, `{"addressComponents": [
			{"longText": "Toronto", "types": ["locality", "political"]},
			{"longText": "Canada", "shortText": "CA", "types": ["country", "political"]}
		]}`)
	})

	if _, err := client.PlaceDetails(context.Background(), placeParams()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet || gotPath != "/v1/places/place-1" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}

	if gotKey != testAPIKey {
		t.Errorf("api key header = %q", gotKey)
	}

	if gotMask != "addressComponents" {
		t.Errorf("field mask = %q", gotMask)
	}

	if gotToken != "token-1" {
		t.Errorf("sessionToken = %q", gotToken)
	}

	if gotLang != "en" {
		t.Errorf("languageCode = %q, want en", gotLang)
	}
}

func TestPlaceDetailsExtractsCityAndCountry(t *testing.T) {
	tests := map[string]struct {
		body string
		want models.City
	}{
		"locality": {
			body: `{"addressComponents": [
				{"longText": "Toronto", "types": ["locality", "political"]},
				{"longText": "Ontario", "types": ["administrative_area_level_1", "political"]},
				{"longText": "Canada", "shortText": "CA", "types": ["country", "political"]}
			]}`,
			want: models.City{City: "Toronto", CountryCode: "CA"},
		},
		"postal town fallback": {
			body: `{"addressComponents": [
				{"longText": "London", "types": ["postal_town"]},
				{"longText": "United Kingdom", "shortText": "GB", "types": ["country", "political"]}
			]}`,
			want: models.City{City: "London", CountryCode: "GB"},
		},
		"level 3 fallback": {
			body: `{"addressComponents": [
				{"longText": "Some Town", "types": ["administrative_area_level_3"]},
				{"longText": "Italy", "shortText": "IT", "types": ["country"]}
			]}`,
			want: models.City{City: "Some Town", CountryCode: "IT"},
		},
		"locality wins over fallbacks": {
			body: `{"addressComponents": [
				{"longText": "Town Level 3", "types": ["administrative_area_level_3"]},
				{"longText": "Postal Town", "types": ["postal_town"]},
				{"longText": "Real City", "types": ["locality"]},
				{"longText": "United States", "shortText": "us", "types": ["country"]}
			]}`,
			want: models.City{City: "Real City", CountryCode: "US"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, tc.body)
			})

			got, err := client.PlaceDetails(context.Background(), placeParams())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestPlaceDetailsIncompletePlaceIsNotFound(t *testing.T) {
	tests := map[string]string{
		"no components": `{}`,
		"no city":       `{"addressComponents": [{"longText": "Canada", "shortText": "CA", "types": ["country"]}]}`,
		"no country":    `{"addressComponents": [{"longText": "Toronto", "types": ["locality"]}]}`,
		"bad country":   `{"addressComponents": [{"longText": "Toronto", "types": ["locality"]}, {"shortText": "CAN", "types": ["country"]}]}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, body)
			})

			_, err := client.PlaceDetails(context.Background(), placeParams())
			requireAPIError(t, err, http.StatusNotFound, MsgCityNotFound)
		})
	}
}

func TestPlaceDetailsGoogleClientErrorsAreNotFound(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusBadRequest} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			})

			_, err := client.PlaceDetails(context.Background(), placeParams())
			requireAPIError(t, err, http.StatusNotFound, MsgCityNotFound)
		})
	}
}

func TestPlaceDetailsProviderFailures(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"server error": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		"rate limited": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		},
		"forbidden key": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		},
		"invalid json": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, `{not json`)
		},
	}

	for name, handler := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, handler)

			_, err := client.PlaceDetails(context.Background(), placeParams())
			requireAPIError(t, err, http.StatusBadGateway, MsgProviderFailed)
		})
	}
}

func TestPlaceDetailsEscapesPlaceID(t *testing.T) {
	var gotPath string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()

		writeJSON(t, w, `{}`)
	})

	params := models.PlaceDetailsParams{PlaceID: "a/b?c", SessionToken: "token-1"}

	if _, err := client.PlaceDetails(context.Background(), params); err == nil {
		t.Fatal("expected not-found error for empty body")
	}

	if gotPath != "/v1/places/a%2Fb%3Fc" {
		t.Fatalf("place ID was not escaped, path = %q", gotPath)
	}
}
