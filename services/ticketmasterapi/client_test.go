package ticketmasterapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"event-explorer/models"
)

const testAPIKey = "secret-test-key"

const fullEventJSON = `{
	"name": "Toronto Raptors vs Celtics",
	"id": "vv1",
	"url": "https://www.ticketmaster.ca/event/vv1",
	"info": "Doors open at 6pm.",
	"pleaseNote": "No bags.",
	"images": [
		{"ratio": "16_9", "url": "https://img/small.jpg", "width": 305},
		{"ratio": "16_9", "url": "https://img/medium.jpg", "width": 640},
		{"ratio": "16_9", "url": "https://img/large.jpg", "width": 1024},
		{"ratio": "3_2", "url": "https://img/3x2.jpg", "width": 640}
	],
	"dates": {"start": {"localDate": "2026-11-05", "localTime": "19:30:00"}},
	"_embedded": {"venues": [
		{"name": "Scotiabank Arena", "city": {"name": "Toronto"}, "address": {"line1": "40 Bay St"}}
	]}
}`

// newTestClient returns a client that talks to a fake Ticketmaster server.
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

func asAPIError(t *testing.T, err error) models.APIError {
	t.Helper()

	apiErr, ok := errors.AsType[models.APIError](err)
	if !ok {
		t.Fatalf("expected models.APIError, got %T: %v", err, err)
	}

	return apiErr
}

func requireAPIError(t *testing.T, err error, status int, message string) {
	t.Helper()

	apiErr := asAPIError(t, err)

	if apiErr.StatusCode != status || apiErr.Message != message {
		t.Fatalf("got %d %q, want %d %q", apiErr.StatusCode, apiErr.Message, status, message)
	}
}

func searchParams() models.EventSearchParams {
	return models.EventSearchParams{City: "Toronto", CountryCode: "CA", Category: models.CategorySports}
}

func TestEventsSendsExpectedRequest(t *testing.T) {
	var gotMethod, gotPath string

	var gotQuery map[string]string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotQuery = map[string]string{}

		for key := range r.URL.Query() {
			gotQuery[key] = r.URL.Query().Get(key)
		}

		writeJSON(t, w, `{}`)
	})

	params := models.EventSearchParams{City: "São Paulo", CountryCode: "BR", Category: models.CategoryMusic}

	if _, err := client.Events(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet || gotPath != "/events.json" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}

	want := map[string]string{
		"apikey":             testAPIKey,
		"city":               "São Paulo",
		"countryCode":        "BR",
		"classificationName": models.CategoryMusic,
		"size":               "6",
	}

	if len(gotQuery) != len(want) {
		t.Errorf("query = %v, want %v", gotQuery, want)
	}

	for key, value := range want {
		if gotQuery[key] != value {
			t.Errorf("query %s = %q, want %q", key, gotQuery[key], value)
		}
	}
}

func TestEventsMapsFields(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"_embedded": {"events": [`+fullEventJSON+`]}}`)
	})

	got, err := client.Events(context.Background(), searchParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := models.Event{
		ID:          "vv1",
		Name:        "Toronto Raptors vs Celtics",
		ImageURL:    "https://img/medium.jpg",
		Date:        "2026-11-05",
		Time:        "19:30:00",
		Venue:       "Scotiabank Arena",
		Address:     "40 Bay St",
		City:        "Toronto",
		Description: "Doors open at 6pm.",
		TicketURL:   "https://www.ticketmaster.ca/event/vv1",
	}

	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %+v, want [%+v]", got, want)
	}
}

func TestEventsKeepsEventsWithMissingOptionalFields(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"_embedded": {"events": [{"id": "e1", "name": "Bare event"}]}}`)
	})

	got, err := client.Events(context.Background(), searchParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 || got[0] != (models.Event{ID: "e1", Name: "Bare event"}) {
		t.Fatalf("got %+v", got)
	}
}

func TestEventsSkipsInvalidEntries(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"_embedded": {"events": [
			{"id": "", "name": "No ID"},
			{"id": "e2", "name": ""},
			{"id": "e3", "name": "Valid"}
		]}}`)
	})

	got, err := client.Events(context.Background(), searchParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 || got[0].ID != "e3" {
		t.Fatalf("got %+v, want only e3", got)
	}
}

func TestEventsCapsResults(t *testing.T) {
	var items []string

	for i := range EventsPageSize + 4 {
		id := "e" + strconv.Itoa(i)
		items = append(items, `{"id": "`+id+`", "name": "Event `+id+`"}`)
	}

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"_embedded": {"events": [`+strings.Join(items, ",")+`]}}`)
	})

	got, err := client.Events(context.Background(), searchParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != EventsPageSize {
		t.Fatalf("got %d events, want %d", len(got), EventsPageSize)
	}
}

func TestEventsNoResultsIsEmptyList(t *testing.T) {
	// Ticketmaster omits "_embedded" entirely when nothing matches.
	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{"page": {"size": 6, "totalElements": 0, "totalPages": 0, "number": 0}}`)
	})

	got, err := client.Events(context.Background(), searchParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got == nil || len(got) != 0 {
		t.Fatalf("want empty non-nil slice, got %#v", got)
	}
}

func TestEventsProviderFailures(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"server error": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		"invalid key": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		},
		"rate limited": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		},
		"invalid json": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, `{not json`)
		},
	}

	for name, handler := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, handler)

			_, err := client.Events(context.Background(), searchParams())
			requireAPIError(t, err, http.StatusBadGateway, MsgProviderFailed)
		})
	}
}

func TestEventsTimeoutIsProviderFailure(t *testing.T) {
	release := make(chan struct{})
	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		<-release
	})
	client.HTTPClient.Timeout = 20 * time.Millisecond

	t.Cleanup(func() { close(release) })

	_, err := client.Events(context.Background(), searchParams())
	requireAPIError(t, err, http.StatusBadGateway, MsgProviderFailed)
}

func TestEventsCanceledRequestIsNotProviderFailure(t *testing.T) {
	release := make(chan struct{})
	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		<-release
	})

	t.Cleanup(func() { close(release) })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Events(ctx, searchParams())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestTransportErrorDoesNotLeakAPIKey(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	client := NewClient(testAPIKey)
	client.BaseURL = server.URL

	// Nothing is listening any more, so the request fails at the transport.
	server.Close()

	_, eventsErr := client.Events(context.Background(), searchParams())
	_, eventErr := client.Event(context.Background(), "vv1")

	for name, err := range map[string]error{"Events": eventsErr, "Event": eventErr} {
		requireAPIError(t, err, http.StatusBadGateway, MsgProviderFailed)

		apiErr := asAPIError(t, err)

		if apiErr.Err == nil {
			t.Fatalf("%s: expected the underlying error to be kept for logging", name)
		}

		if strings.Contains(apiErr.Err.Error(), testAPIKey) || strings.Contains(apiErr.Error(), testAPIKey) {
			t.Errorf("%s: API key leaked into error: %v", name, apiErr.Err)
		}
	}
}

func TestEventSendsExpectedRequest(t *testing.T) {
	var gotMethod, gotPath, gotKey string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.EscapedPath()
		gotKey = r.URL.Query().Get("apikey")

		writeJSON(t, w, fullEventJSON)
	})

	got, err := client.Event(context.Background(), "vv1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodGet || gotPath != "/events/vv1.json" {
		t.Errorf("got %s %s", gotMethod, gotPath)
	}

	if gotKey != testAPIKey {
		t.Errorf("apikey = %q", gotKey)
	}

	if got.ID != "vv1" || got.TicketURL != "https://www.ticketmaster.ca/event/vv1" {
		t.Errorf("unexpected event: %+v", got)
	}
}

func TestEventEscapesEventID(t *testing.T) {
	var gotPath string

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()

		writeJSON(t, w, `{}`)
	})

	if _, err := client.Event(context.Background(), "a/b?c"); err == nil {
		t.Fatal("expected not-found error for empty body")
	}

	if gotPath != "/events/a%2Fb%3Fc.json" {
		t.Fatalf("event ID was not escaped, path = %q", gotPath)
	}
}

func TestEventClientErrorsAreNotFound(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusBadRequest} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			})

			_, err := client.Event(context.Background(), "missing")
			requireAPIError(t, err, http.StatusNotFound, MsgEventNotFound)
		})
	}
}

func TestEventInvalidPayloadIsNotFound(t *testing.T) {
	tests := map[string]string{
		"empty object": `{}`,
		"no id":        `{"name": "No ID"}`,
		"no name":      `{"id": "e1"}`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, body)
			})

			_, err := client.Event(context.Background(), "e1")
			requireAPIError(t, err, http.StatusNotFound, MsgEventNotFound)
		})
	}
}

func TestEventProviderFailures(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"server error": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
		"invalid key": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		},
		"rate limited": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		},
		"invalid json": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, `{not json`)
		},
	}

	for name, handler := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, handler)

			_, err := client.Event(context.Background(), "e1")
			requireAPIError(t, err, http.StatusBadGateway, MsgProviderFailed)
		})
	}
}

func TestDescriptionFallbackOrder(t *testing.T) {
	tests := map[string]struct {
		payload string
		want    string
	}{
		"info first":       {`"info": "I", "pleaseNote": "P", "description": "D"`, "I"},
		"then note":        {`"pleaseNote": "P", "description": "D"`, "P"},
		"then description": {`"description": "D"`, "D"},
		"none":             {`"url": "https://x"`, ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(t, w, `{"id": "e1", "name": "N", `+tc.payload+`}`)
			})

			got, err := client.Event(context.Background(), "e1")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Description != tc.want {
				t.Fatalf("description = %q, want %q", got.Description, tc.want)
			}
		})
	}
}

func TestPickImage(t *testing.T) {
	type image = struct {
		Ratio string `json:"ratio"`
		URL   string `json:"url"`
		Width int    `json:"width"`
	}

	tests := map[string]struct {
		images []image
		want   string
	}{
		"smallest 16_9 that is wide enough": {
			images: []image{
				{"16_9", "small", 305},
				{"16_9", "large", 1024},
				{"16_9", "medium", 640},
				{"3_2", "other-ratio", 640},
			},
			want: "medium",
		},
		"widest 16_9 when all are too small": {
			images: []image{{"16_9", "tiny", 100}, {"16_9", "small", 305}},
			want:   "small",
		},
		"first image when no 16_9": {
			images: []image{{"3_2", "first", 640}, {"4_3", "second", 800}},
			want:   "first",
		},
		"skips images without a url": {
			images: []image{{"16_9", "", 640}, {"16_9", "real", 1024}},
			want:   "real",
		},
		"no images": {images: nil, want: ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			payload := eventPayload{Images: tc.images}

			if got := payload.pickImage(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
