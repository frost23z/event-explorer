package models

const (
	MinAutocompleteRunes      = 3
	MaxAutocompleteInputBytes = 200
	MsgInvalidSearch          = "Use a short city name and a valid search session."
	MsgInvalidPlace           = "Use a valid place and search session."
)

type AutocompleteParams struct {
	Input        string
	SessionToken string
}

type Suggestion struct {
	PlaceID string `json:"placeId"`
	Text    string `json:"text"`
}

type AutocompleteResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
}

type PlaceDetailsParams struct {
	PlaceID      string
	SessionToken string
}

type City struct {
	City        string `json:"city"`
	CountryCode string `json:"countryCode"`
}
