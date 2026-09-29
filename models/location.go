package models

const (
	MaxAutocompleteInputBytes = 200
	MsgInvalidSearch          = "Use a short city name and a valid search session."
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
