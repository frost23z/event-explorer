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

var sessionTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,36}$`)

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

func hasControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}
