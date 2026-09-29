package ticketmasterapi

const (
	EventsPageSize = 6

	MsgProviderFailed = "Unable to retrieve events. Please try again."
	MsgEventNotFound  = "Event not found."

	// Ratio and minimum width used when picking a card image.
	imageRatio    = "16_9"
	imageMinWidth = 640
)

type eventsResponse struct {
	Embedded struct {
		Events []eventPayload `json:"events"`
	} `json:"_embedded"`
}

type eventPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Info        string `json:"info"`
	PleaseNote  string `json:"pleaseNote"`
	Images      []struct {
		Ratio string `json:"ratio"`
		URL   string `json:"url"`
		Width int    `json:"width"`
	} `json:"images"`
	Dates struct {
		Start struct {
			LocalDate string `json:"localDate"`
			LocalTime string `json:"localTime"`
		} `json:"start"`
	} `json:"dates"`
	Embedded struct {
		Venues []struct {
			Name string `json:"name"`
			City struct {
				Name string `json:"name"`
			} `json:"city"`
			Address struct {
				Line1 string `json:"line1"`
			} `json:"address"`
		} `json:"venues"`
	} `json:"_embedded"`
}
