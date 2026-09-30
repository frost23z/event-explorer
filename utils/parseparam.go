package utils

import (
	"event-explorer/models"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	sessionTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,36}$`)
	placeIDPattern      = regexp.MustCompile(`^[A-Za-z0-9_-]{1,200}$`)
	countryCodePattern  = regexp.MustCompile(`^[A-Za-z]{2}$`)
	eventIDPattern      = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
)

func ParseAutocompleteParams(query url.Values) (models.AutocompleteParams, error) {
	sessionToken := strings.TrimSpace(query.Get("sessionToken"))
	input := strings.TrimSpace(query.Get("input"))

	if input == "" ||
		!sessionTokenPattern.MatchString(sessionToken) ||
		len(input) > models.MaxAutocompleteInputBytes ||
		!utf8.ValidString(input) ||
		hasControlChars(input) {
		return models.AutocompleteParams{}, models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    models.MsgInvalidSearch,
		}
	}

	return models.AutocompleteParams{Input: input, SessionToken: sessionToken}, nil
}

func ParsePlaceDetailsParams(placeID string, query url.Values) (models.PlaceDetailsParams, error) {
	sessionToken := strings.TrimSpace(query.Get("sessionToken"))

	if !placeIDPattern.MatchString(placeID) || !sessionTokenPattern.MatchString(sessionToken) {
		return models.PlaceDetailsParams{}, models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    models.MsgInvalidPlace,
		}
	}

	return models.PlaceDetailsParams{PlaceID: placeID, SessionToken: sessionToken}, nil
}

func ParseEventSearchParams(query url.Values) (models.EventSearchParams, error) {
	city := strings.ToLower(strings.TrimSpace(query.Get("city")))
	countryCode := strings.ToUpper(strings.TrimSpace(query.Get("countryCode")))
	category := strings.ToLower(strings.TrimSpace(query.Get("category")))

	if city == "" ||
		len(city) > models.MaxCityBytes ||
		!utf8.ValidString(city) ||
		hasControlChars(city) ||
		!countryCodePattern.MatchString(countryCode) ||
		(category != models.CategoryMusic && category != models.CategorySports) {
		return models.EventSearchParams{}, models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    models.MsgInvalidEventSearch,
		}
	}

	return models.EventSearchParams{City: city, CountryCode: countryCode, Category: category}, nil
}

func ParseEventID(eventID string) (string, error) {
	if !eventIDPattern.MatchString(eventID) {
		return "", models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    models.MsgInvalidEvent,
		}
	}

	return eventID, nil
}

func ParseListingParams(query url.Values) (models.EventSearchParams, error) {
	rawCity := strings.TrimSpace(query.Get("city"))
	city := strings.ToLower(rawCity)
	countryCode := strings.ToUpper(strings.TrimSpace(query.Get("countryCode")))

	if city == "" ||
		len(city) > models.MaxCityBytes ||
		!utf8.ValidString(rawCity) || // before ToLower, which would replace bad bytes
		hasControlChars(city) ||
		!countryCodePattern.MatchString(countryCode) {
		return models.EventSearchParams{}, models.APIError{
			StatusCode: http.StatusBadRequest,
			Message:    models.MsgInvalidListing,
		}
	}

	return models.EventSearchParams{City: city, CountryCode: countryCode}, nil
}

func hasControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
