package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"event-explorer/models"
	"event-explorer/services/ticketmasterapi"
)

// newTestEventService returns a service backed by a fake Ticketmaster server
// and a counter of the requests that reached that server.
func newTestEventService(t *testing.T, handler http.HandlerFunc) (*EventService, *atomic.Int32) {
	t.Helper()

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		handler(w, r)
	}))
	t.Cleanup(server.Close)

	client := ticketmasterapi.NewClient("test-key")
	client.BaseURL = server.URL

	return NewEventService(client), &calls
}

func writeJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	// nolint:gosec
	if _, err := w.Write([]byte(body)); err != nil {
		t.Errorf("write response: %v", err)
	}
}

func requireAPIError(t *testing.T, err error, status int) {
	t.Helper()

	apiErr, ok := errors.AsType[models.APIError](err)
	if !ok {
		t.Fatalf("expected models.APIError, got %T: %v", err, err)
	}

	if apiErr.StatusCode != status {
		t.Fatalf("status = %d, want %d", apiErr.StatusCode, status)
	}
}

// categoryHandler serves one event per category ("music-1", "sports-1"); a
// category listed in failing answers 500 instead.
func categoryHandler(t *testing.T, failing ...string) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		category := r.URL.Query().Get("classificationName")

		for _, f := range failing {
			if f == category {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		writeJSON(t, w, `{"_embedded": {"events": [{"id": "`+category+`-1", "name": "`+category+` event"}]}}`)
	}
}

func musicSearch() models.EventSearchParams {
	return models.EventSearchParams{City: "toronto", CountryCode: "CA", Category: models.CategoryMusic}
}

func TestEventsSuccess(t *testing.T) {
	service, calls := newTestEventService(t, categoryHandler(t))

	got, err := service.Events(context.Background(), musicSearch())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 || got[0].ID != "music-1" {
		t.Fatalf("got %+v", got)
	}

	if calls.Load() != 1 {
		t.Fatalf("API calls = %d, want 1", calls.Load())
	}
}

func TestEventsSendsSearchToTicketmaster(t *testing.T) {
	var gotCity, gotCountry, gotCategory string

	service, _ := newTestEventService(t, func(w http.ResponseWriter, r *http.Request) {
		gotCity = r.URL.Query().Get("city")
		gotCountry = r.URL.Query().Get("countryCode")
		gotCategory = r.URL.Query().Get("classificationName")

		writeJSON(t, w, `{}`)
	})

	if _, err := service.Events(context.Background(), musicSearch()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotCity != "toronto" || gotCountry != "CA" || gotCategory != models.CategoryMusic {
		t.Fatalf("Ticketmaster got city=%q countryCode=%q classificationName=%q", gotCity, gotCountry, gotCategory)
	}
}

func TestEventsAreCachedOnSuccess(t *testing.T) {
	service, calls := newTestEventService(t, categoryHandler(t))

	for range 3 {
		got, err := service.Events(context.Background(), musicSearch())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 || got[0].ID != "music-1" {
			t.Fatalf("got %+v", got)
		}
	}

	if calls.Load() != 1 {
		t.Fatalf("API calls = %d, want 1: the second and third lookups must be cache hits", calls.Load())
	}
}

func TestClearCacheForcesANewLookup(t *testing.T) {
	service, calls := newTestEventService(t, categoryHandler(t))

	for range 2 {
		if _, err := service.Events(context.Background(), musicSearch()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if calls.Load() != 1 {
		t.Fatalf("API calls = %d, want 1 before the clear", calls.Load())
	}

	if got := service.ClearCache(); got != 1 {
		t.Fatalf("cleared %d entries, want 1", got)
	}

	if _, err := service.Events(context.Background(), musicSearch()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls.Load() != 2 {
		t.Fatalf("API calls = %d, want 2: the lookup after the clear must reach Ticketmaster", calls.Load())
	}
}

func TestEmptyListIsCached(t *testing.T) {
	service, calls := newTestEventService(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(t, w, `{}`)
	})

	for range 2 {
		got, err := service.Events(context.Background(), musicSearch())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 0 {
			t.Fatalf("got %+v, want no events", got)
		}
	}

	if calls.Load() != 1 {
		t.Fatalf("API calls = %d, want 1: an empty list is a successful result", calls.Load())
	}
}

func TestDifferentSearchesAreCachedSeparately(t *testing.T) {
	service, calls := newTestEventService(t, categoryHandler(t))

	searches := []models.EventSearchParams{
		musicSearch(),
		{City: "toronto", CountryCode: "CA", Category: models.CategorySports},
		{City: "ottawa", CountryCode: "CA", Category: models.CategoryMusic},
		{City: "toronto", CountryCode: "US", Category: models.CategoryMusic},
	}

	for _, search := range searches {
		if _, err := service.Events(context.Background(), search); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if int(calls.Load()) != len(searches) {
		t.Fatalf("API calls = %d, want %d (one per distinct city/country/category)", calls.Load(), len(searches))
	}
}

func TestExpiredCacheEntryIsRefreshed(t *testing.T) {
	service, calls := newTestEventService(t, categoryHandler(t))

	now := time.Date(2026, time.October, 1, 12, 0, 0, 0, time.UTC)
	service.cache.now = func() time.Time { return now }

	lookup := func() {
		t.Helper()

		if _, err := service.Events(context.Background(), musicSearch()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	lookup()

	now = now.Add(cacheTTL - time.Second)

	lookup()

	if calls.Load() != 1 {
		t.Fatalf("API calls = %d, want 1 before expiry", calls.Load())
	}

	now = now.Add(2 * time.Second)

	lookup()

	if calls.Load() != 2 {
		t.Fatalf("API calls = %d, want 2: an expired entry must be fetched again", calls.Load())
	}

	// The refreshed entry is cached again.
	lookup()

	if calls.Load() != 2 {
		t.Fatalf("API calls = %d, want 2 after the refresh", calls.Load())
	}
}

func TestEventsAPIFailure(t *testing.T) {
	service, _ := newTestEventService(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	got, err := service.Events(context.Background(), musicSearch())
	requireAPIError(t, err, http.StatusBadGateway)

	if got != nil {
		t.Fatalf("got %+v with an error", got)
	}
}

func TestFailedLookupIsNotCached(t *testing.T) {
	var requests atomic.Int32

	service, calls := newTestEventService(t, func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		categoryHandler(t)(w, r)
	})

	if _, err := service.Events(context.Background(), musicSearch()); err == nil {
		t.Fatal("expected the first lookup to fail")
	}

	got, err := service.Events(context.Background(), musicSearch())
	if err != nil {
		t.Fatalf("the retry should reach Ticketmaster again and succeed: %v", err)
	}

	if len(got) != 1 || calls.Load() != 2 {
		t.Fatalf("got %+v after %d API calls, want the events after 2", got, calls.Load())
	}
}

func TestCanceledLookupIsNotCached(t *testing.T) {
	service, calls := newTestEventService(t, categoryHandler(t))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := service.Events(ctx, musicSearch()); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	if _, err := service.Events(context.Background(), musicSearch()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls.Load() != 1 {
		t.Fatalf("API calls = %d, want 1: the canceled lookup must not have stored anything", calls.Load())
	}
}

func TestEventSuccess(t *testing.T) {
	service, _ := newTestEventService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/e1.json" {
			t.Errorf("path = %q", r.URL.Path)
		}

		writeJSON(t, w, `{"id": "e1", "name": "Concert", "url": "https://www.ticketmaster.com/e1"}`)
	})

	got, err := service.Event(context.Background(), "e1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != "e1" || got.Name != "Concert" || got.TicketURL != "https://www.ticketmaster.com/e1" {
		t.Fatalf("got %+v", got)
	}
}

func TestEventInvalid(t *testing.T) {
	tests := map[string]http.HandlerFunc{
		"unknown event": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		},
		"malformed id rejected by Ticketmaster": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		},
		"payload without id": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, `{"name": "No ID"}`)
		},
		"payload without name": func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(t, w, `{"id": "e1"}`)
		},
	}

	for name, handler := range tests {
		t.Run(name, func(t *testing.T) {
			service, _ := newTestEventService(t, handler)

			_, err := service.Event(context.Background(), "missing")
			requireAPIError(t, err, http.StatusNotFound)
		})
	}
}

func TestEventAPIFailure(t *testing.T) {
	service, _ := newTestEventService(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := service.Event(context.Background(), "e1")
	requireAPIError(t, err, http.StatusBadGateway)
}

func TestMusicAndSportsSuccess(t *testing.T) {
	service, calls := newTestEventService(t, categoryHandler(t))

	music, sports := service.MusicAndSports(context.Background(), "toronto", "CA")

	if music.Err != nil || len(music.Events) != 1 || music.Events[0].ID != "music-1" {
		t.Errorf("music = %+v", music)
	}

	if sports.Err != nil || len(sports.Events) != 1 || sports.Events[0].ID != "sports-1" {
		t.Errorf("sports = %+v", sports)
	}

	if calls.Load() != 2 {
		t.Errorf("API calls = %d, want one request per category", calls.Load())
	}
}

func TestMusicAndSportsKeepsSuccessfulSectionWhenOtherFails(t *testing.T) {
	for _, failing := range []string{models.CategoryMusic, models.CategorySports} {
		t.Run(failing+" fails", func(t *testing.T) {
			service, _ := newTestEventService(t, categoryHandler(t, failing))

			music, sports := service.MusicAndSports(context.Background(), "toronto", "CA")

			results := map[string]models.CategoryResult{models.CategoryMusic: music, models.CategorySports: sports}

			for category, result := range results {
				if category == failing {
					requireAPIError(t, result.Err, http.StatusBadGateway)
					continue
				}

				if result.Err != nil || len(result.Events) != 1 {
					t.Errorf("%s: want its events kept, got %+v", category, result)
				}
			}
		})
	}
}

func TestMusicAndSportsBothFail(t *testing.T) {
	service, _ := newTestEventService(t, categoryHandler(t, models.CategoryMusic, models.CategorySports))

	music, sports := service.MusicAndSports(context.Background(), "toronto", "CA")

	requireAPIError(t, music.Err, http.StatusBadGateway)
	requireAPIError(t, sports.Err, http.StatusBadGateway)
}

func TestMusicAndSportsRunConcurrently(t *testing.T) {
	var inFlight, maxInFlight atomic.Int32

	service, _ := newTestEventService(t, func(w http.ResponseWriter, r *http.Request) {
		current := inFlight.Add(1)

		for {
			highest := maxInFlight.Load()
			if current <= highest || maxInFlight.CompareAndSwap(highest, current) {
				break
			}
		}

		// Stay in flight long enough for the other category's request to arrive.
		time.Sleep(150 * time.Millisecond)
		inFlight.Add(-1)

		categoryHandler(t)(w, r)
	})

	service.MusicAndSports(context.Background(), "toronto", "CA")

	if maxInFlight.Load() != 2 {
		t.Fatalf("at most %d requests were in flight at once; both categories must be requested together", maxInFlight.Load())
	}
}

func TestMusicAndSportsReportsCacheHits(t *testing.T) {
	service, _ := newTestEventService(t, categoryHandler(t, "sports"))

	music, sports := service.MusicAndSports(context.Background(), "toronto", "CA")
	if music.Cached || sports.Cached {
		t.Fatalf("first lookup: music cached = %v, sports cached = %v; want both false", music.Cached, sports.Cached)
	}

	music, sports = service.MusicAndSports(context.Background(), "toronto", "CA")
	if !music.Cached {
		t.Error("second lookup: the successful Music section must come from the cache")
	}

	if sports.Cached {
		t.Error("second lookup: the failed Sports section was never cached")
	}
}

func TestMusicAndSportsUsesTheCache(t *testing.T) {
	service, calls := newTestEventService(t, categoryHandler(t))

	service.MusicAndSports(context.Background(), "toronto", "CA")
	service.MusicAndSports(context.Background(), "toronto", "CA")

	if calls.Load() != 2 {
		t.Fatalf("API calls = %d, want 2: the second page view must come from the cache", calls.Load())
	}
}

func TestMusicAndSportsCachesOnlyTheSuccessfulSection(t *testing.T) {
	var sportsFails atomic.Bool

	sportsFails.Store(true)

	service, calls := newTestEventService(t, func(w http.ResponseWriter, r *http.Request) {
		if sportsFails.Load() && r.URL.Query().Get("classificationName") == models.CategorySports {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		categoryHandler(t)(w, r)
	})

	service.MusicAndSports(context.Background(), "toronto", "CA")

	sportsFails.Store(false)

	music, sports := service.MusicAndSports(context.Background(), "toronto", "CA")

	if music.Err != nil || sports.Err != nil || len(sports.Events) != 1 {
		t.Fatalf("music = %+v, sports = %+v", music, sports)
	}

	// 2 first time + only sports again (music was cached).
	if calls.Load() != 3 {
		t.Fatalf("API calls = %d, want 3", calls.Load())
	}
}
