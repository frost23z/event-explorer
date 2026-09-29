package googleplacesapi

const MaxAutocompleteSuggestions = 5
const MsgProviderFailed = "Unable to retrieve places. Please try again."
const MsgCityNotFound = "City not found."

type autocompleteRequest struct {
	Input                string   `json:"input"`
	IncludedPrimaryTypes []string `json:"includedPrimaryTypes"`
	SessionToken         string   `json:"sessionToken"`
	LanguageCode         string   `json:"languageCode"`
}

type autocompleteResponse struct {
	Suggestions []struct {
		PlacePrediction *struct {
			PlaceID string `json:"placeId"`
			Text    struct {
				Text string `json:"text"`
			} `json:"text"`
		} `json:"placePrediction"`
	} `json:"suggestions"`
}

type placeResponse struct {
	AddressComponents []struct {
		LongText  string   `json:"longText"`
		ShortText string   `json:"shortText"`
		Types     []string `json:"types"`
	} `json:"addressComponents"`
}
