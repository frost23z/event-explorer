package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"event-explorer/controllers"
	"event-explorer/services"
	"event-explorer/services/ticketmasterapi"

	beego "github.com/beego/beego/v2/server/web"
)

func TestListingReportsCacheHeaders(t *testing.T) {
	ticketmaster := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"_embedded": {"events": [{"id": "e1", "name": "Some event"}]}}`)
	}))
	t.Cleanup(ticketmaster.Close)

	client := ticketmasterapi.NewClient("test-key")
	client.BaseURL = ticketmaster.URL
	controllers.SetEventService(services.NewEventService(client))

	get := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events?city=Toronto&countryCode=CA", nil)
		beego.BeeApp.Handlers.ServeHTTP(w, r)

		return w
	}

	for i, want := range []string{"MISS", "HIT"} {
		w := get()

		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200", i+1, w.Code)
		}

		for _, header := range []string{"X-Music-Cache", "X-Sports-Cache"} {
			if got := w.Header().Get(header); got != want {
				t.Errorf("request %d: %s = %q, want %q", i+1, header, got, want)
			}
		}
	}
}
