package utils

import (
	"errors"
	"net/http"
	"testing"

	"event-explorer/models"
)

func TestValidateTicketURLAcceptsApprovedHosts(t *testing.T) {
	tests := map[string]string{
		"ticketmaster.com":           "https://www.ticketmaster.com/event/abc",
		"country domain":             "https://www.ticketmaster.ca/event/abc?tm_link=1",
		"two-part country domain":    "https://www.ticketmaster.co.uk/event/abc",
		"bare approved host":         "https://ticketmaster.com/event/abc",
		"nested subdomain":           "https://a.b.ticketmaster.com/event/abc",
		"affiliate redirect":         "https://ticketmaster.evyy.net/c/1/2/3?u=https%3A%2F%2Fwww.ticketmaster.com%2Fevent%2Fabc",
		"live nation":                "https://www.livenation.com/event/abc",
		"uppercase host":             "https://WWW.TICKETMASTER.COM/event/abc",
		"explicit https port":        "https://www.ticketmaster.com:443/event/abc",
		"fragment and query kept":    "https://www.ticketmaster.com/event/abc?a=1&b=2#tickets",
		"encoded characters in path": "https://www.ticketmaster.com/event/a%20b",
	}

	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ValidateTicketURL(raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != raw {
				t.Fatalf("got %q, want %q", got, raw)
			}
		})
	}
}

func TestValidateTicketURLRejects(t *testing.T) {
	// nolint:gosec
	tests := map[string]string{
		"empty":                  "",
		"blank":                  "   ",
		"relative path":          "/redirect/other",
		"no scheme":              "www.ticketmaster.com/event/abc",
		"protocol relative":      "//www.ticketmaster.com/event/abc",
		"http":                   "http://www.ticketmaster.com/event/abc",
		"javascript":             "javascript:alert(1)",
		"data":                   "data:text/html,<script>alert(1)</script>",
		"ftp":                    "ftp://www.ticketmaster.com/event/abc",
		"no host":                "https://",
		"unknown host":           "https://example.com/event/abc",
		"approved host in path":  "https://evil.com/www.ticketmaster.com",
		"approved host in query": "https://evil.com/?u=https://www.ticketmaster.com",
		"approved host as label": "https://ticketmaster.com.evil.com/event/abc",
		"suffix lookalike":       "https://evilticketmaster.com/event/abc",
		"prefix lookalike":       "https://ticketmaster.com-tickets.com/event/abc",
		"unrelated tld":          "https://www.ticketmaster.evil/event/abc",
		"userinfo trick":         "https://www.ticketmaster.com@evil.com/event/abc",
		"userinfo on real host":  "https://user:pass@www.ticketmaster.com/event/abc",
		"backslash trick":        "https://evil.com\\@www.ticketmaster.com/event/abc",
		"unicode lookalike":      "https://www.ticketmaster.cоm/event/abc", // Cyrillic "о"
		"space in host":          "https://www.ticket master.com/event/abc",
		"control character":      "https://www.ticketmaster.com/event\n/abc",
	}

	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ValidateTicketURL(raw)

			apiErr, ok := errors.AsType[models.APIError](err)
			if !ok {
				t.Fatalf("expected models.APIError for %q, got %T: %v (result %q)", raw, err, err, got)
			}

			if apiErr.StatusCode != http.StatusBadRequest || apiErr.Message != models.MsgTicketUnavailable {
				t.Fatalf("got %d %q, want 400 %q", apiErr.StatusCode, apiErr.Message, models.MsgTicketUnavailable)
			}

			if apiErr.Err == nil {
				t.Error("the reason should be kept for logging")
			}

			if got != "" {
				t.Fatalf("returned %q together with an error", got)
			}
		})
	}
}
