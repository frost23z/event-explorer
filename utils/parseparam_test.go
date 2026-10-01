package utils

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"event-explorer/models"
)

func requireBadRequest(t *testing.T, err error, message string) {
	t.Helper()

	apiErr, ok := errors.AsType[models.APIError](err)
	if !ok {
		t.Fatalf("expected models.APIError, got %T: %v", err, err)
	}

	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Message != message {
		t.Fatalf("got %d %q, want 400 %q", apiErr.StatusCode, apiErr.Message, message)
	}
}

func TestParseAutocompleteParams(t *testing.T) {
	valid := strings.Repeat("a", 36)
	// nolint:gosec
	tests := map[string]struct {
		query   url.Values
		want    models.AutocompleteParams
		wantErr bool
	}{
		"valid": {
			query: url.Values{"input": {"Toronto"}, "sessionToken": {"abc-123_DEF"}},
			want:  models.AutocompleteParams{Input: "Toronto", SessionToken: "abc-123_DEF"},
		},
		"trims whitespace": {
			query: url.Values{"input": {"  Toronto  "}, "sessionToken": {" tok "}},
			want:  models.AutocompleteParams{Input: "Toronto", SessionToken: "tok"},
		},
		"uuid token": {
			query: url.Values{"input": {"Toronto"}, "sessionToken": {"123e4567-e89b-12d3-a456-426614174000"}},
			want:  models.AutocompleteParams{Input: "Toronto", SessionToken: "123e4567-e89b-12d3-a456-426614174000"},
		},
		"max length token": {
			query: url.Values{"input": {"Toronto"}, "sessionToken": {valid}},
			want:  models.AutocompleteParams{Input: "Toronto", SessionToken: valid},
		},
		"unicode input": {
			query: url.Values{"input": {"Zürich"}, "sessionToken": {"tok"}},
			want:  models.AutocompleteParams{Input: "Zürich", SessionToken: "tok"},
		},
		"missing input":       {query: url.Values{"sessionToken": {"tok"}}, wantErr: true},
		"blank input":         {query: url.Values{"input": {"   "}, "sessionToken": {"tok"}}, wantErr: true},
		"missing token":       {query: url.Values{"input": {"Toronto"}}, wantErr: true},
		"token too long":      {query: url.Values{"input": {"Toronto"}, "sessionToken": {valid + "a"}}, wantErr: true},
		"token bad chars":     {query: url.Values{"input": {"Toronto"}, "sessionToken": {"a/b"}}, wantErr: true},
		"input too long":      {query: url.Values{"input": {strings.Repeat("a", models.MaxAutocompleteInputBytes+1)}, "sessionToken": {"tok"}}, wantErr: true},
		"input control chars": {query: url.Values{"input": {"To\nronto"}, "sessionToken": {"tok"}}, wantErr: true},
		"input invalid utf8":  {query: url.Values{"input": {"To\xffronto"}, "sessionToken": {"tok"}}, wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseAutocompleteParams(tc.query)

			if tc.wantErr {
				requireBadRequest(t, err, models.MsgInvalidSearch)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestParsePlaceDetailsParams(t *testing.T) {
	tests := map[string]struct {
		placeID string
		query   url.Values
		want    models.PlaceDetailsParams
		wantErr bool
	}{
		"valid": {
			placeID: "ChIJpTvG15DL1IkRd8S0KlBVNTI",
			query:   url.Values{"sessionToken": {"tok-1"}},
			want:    models.PlaceDetailsParams{PlaceID: "ChIJpTvG15DL1IkRd8S0KlBVNTI", SessionToken: "tok-1"},
		},
		"empty place id":      {placeID: "", query: url.Values{"sessionToken": {"tok"}}, wantErr: true},
		"place id with slash": {placeID: "a/b", query: url.Values{"sessionToken": {"tok"}}, wantErr: true},
		"place id with dots":  {placeID: "..", query: url.Values{"sessionToken": {"tok"}}, wantErr: true},
		"place id too long": {
			placeID: strings.Repeat("a", 201),
			query:   url.Values{"sessionToken": {"tok"}},
			wantErr: true,
		},
		"missing token":  {placeID: "abc", query: url.Values{}, wantErr: true},
		"bad token":      {placeID: "abc", query: url.Values{"sessionToken": {"a b"}}, wantErr: true},
		"token too long": {placeID: "abc", query: url.Values{"sessionToken": {strings.Repeat("a", 37)}}, wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParsePlaceDetailsParams(tc.placeID, tc.query)

			if tc.wantErr {
				requireBadRequest(t, err, models.MsgInvalidPlace)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestParseListingParams(t *testing.T) {
	tests := map[string]struct {
		query   url.Values
		want    models.EventSearchParams
		wantErr bool
	}{
		"valid": {
			query: url.Values{"city": {"Toronto"}, "countryCode": {"CA"}},
			want:  models.EventSearchParams{City: "toronto", CountryCode: "CA"},
		},
		"city is lower-cased for the cache key": {
			query: url.Values{"city": {"ToRoNtO"}, "countryCode": {"CA"}},
			want:  models.EventSearchParams{City: "toronto", CountryCode: "CA"},
		},
		"country code is upper-cased": {
			query: url.Values{"city": {"London"}, "countryCode": {"gb"}},
			want:  models.EventSearchParams{City: "london", CountryCode: "GB"},
		},
		"trims whitespace": {
			query: url.Values{"city": {"  Toronto "}, "countryCode": {" CA "}},
			want:  models.EventSearchParams{City: "toronto", CountryCode: "CA"},
		},
		"unicode city": {
			query: url.Values{"city": {"São Paulo"}, "countryCode": {"BR"}},
			want:  models.EventSearchParams{City: "são paulo", CountryCode: "BR"},
		},
		"max length city": {
			query: url.Values{"city": {strings.Repeat("a", models.MaxCityRunes)}, "countryCode": {"CA"}},
			want:  models.EventSearchParams{City: strings.Repeat("a", models.MaxCityRunes), CountryCode: "CA"},
		},
		"missing city":        {query: url.Values{"countryCode": {"CA"}}, wantErr: true},
		"blank city":          {query: url.Values{"city": {"  "}, "countryCode": {"CA"}}, wantErr: true},
		"missing country":     {query: url.Values{"city": {"Toronto"}}, wantErr: true},
		"country too long":    {query: url.Values{"city": {"Toronto"}, "countryCode": {"CAN"}}, wantErr: true},
		"country too short":   {query: url.Values{"city": {"Toronto"}, "countryCode": {"C"}}, wantErr: true},
		"country not letters": {query: url.Values{"city": {"Toronto"}, "countryCode": {"C1"}}, wantErr: true},
		"city too long": {
			query:   url.Values{"city": {strings.Repeat("a", models.MaxCityRunes+1)}, "countryCode": {"CA"}},
			wantErr: true,
		},
		"city control chars": {query: url.Values{"city": {"To\nronto"}, "countryCode": {"CA"}}, wantErr: true},
		"city invalid utf8":  {query: url.Values{"city": {"To\xffronto"}, "countryCode": {"CA"}}, wantErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseListingParams(tc.query)

			if tc.wantErr {
				requireBadRequest(t, err, models.MsgInvalidListing)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}

			if got.Category != "" {
				t.Fatalf("category = %q; the listing page asks for both categories", got.Category)
			}
		})
	}
}

func TestParseEventID(t *testing.T) {
	valid := []string{"vvG1IZ4Ok1TLpT", "Z7r9jZ1AdFj7e", "G5diZ9ZL_jCmJ", "a-b_c", strings.Repeat("a", 64)}

	for _, id := range valid {
		got, err := ParseEventID(id)
		if err != nil || got != id {
			t.Errorf("%q: got %q, %v", id, got, err)
		}
	}

	invalid := []string{"", "a/b", "..", "a.json", "a b", "a?b=c", "bad$id", "a%2Fb", strings.Repeat("a", 65)}

	for _, id := range invalid {
		got, err := ParseEventID(id)

		requireBadRequest(t, err, models.MsgInvalidEvent)

		if got != "" {
			t.Errorf("%q: returned %q together with an error", id, got)
		}
	}
}
