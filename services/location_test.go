package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"event-explorer/models"
	"event-explorer/services/googleplacesapi"
)

// newTestService returns a service backed by a fake Google server and a
// counter of how many requests reached that server.
func newTestService(t *testing.T) (*LocationService, *atomic.Int32) {
	t.Helper()

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write([]byte(`{"suggestions": [
			{"placePrediction": {"placeId": "p1", "text": {"text": "Toronto, ON, Canada"}}}
		]}`))
		if err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client := googleplacesapi.NewClient("test-key")
	client.BaseURL = server.URL

	return NewLocationService(client), &calls
}

func TestAutocompleteShortInputSkipsGoogle(t *testing.T) {
	service, calls := newTestService(t)

	for _, input := range []string{"a", "to", "äö"} {
		got, err := service.Autocomplete(context.Background(), models.AutocompleteParams{
			Input:        input,
			SessionToken: "token-1",
		})
		if err != nil {
			t.Fatalf("%q: unexpected error: %v", input, err)
		}

		if got.Suggestions == nil || len(got.Suggestions) != 0 {
			t.Errorf("%q: want empty non-nil suggestions, got %#v", input, got.Suggestions)
		}
	}

	if calls.Load() != 0 {
		t.Fatalf("Google was called %d times for short input, want 0", calls.Load())
	}
}

func TestAutocompleteMinimumLengthCallsGoogle(t *testing.T) {
	service, calls := newTestService(t)

	// Three runes, but more than three bytes: length is counted in runes.
	for _, input := range []string{"tor", "äöü"} {
		got, err := service.Autocomplete(context.Background(), models.AutocompleteParams{
			Input:        input,
			SessionToken: "token-1",
		})
		if err != nil {
			t.Fatalf("%q: unexpected error: %v", input, err)
		}

		if len(got.Suggestions) != 1 {
			t.Errorf("%q: got %d suggestions, want 1", input, len(got.Suggestions))
		}
	}

	if calls.Load() != 2 {
		t.Fatalf("Google was called %d times, want 2", calls.Load())
	}
}
