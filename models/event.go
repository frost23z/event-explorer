package models

const (
	CategoryMusic  = "music"
	CategorySports = "sports"

	MaxCityRunes = 80

	MsgInvalidEvent      = "Use a valid event."
	MsgInvalidListing    = "Use a city and a two-letter country code."
	MsgTicketUnavailable = "Tickets are not available for this event right now."
)

type EventSearchParams struct {
	City        string
	CountryCode string
	Category    string
}

type Event struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ImageURL    string `json:"imageUrl"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Venue       string `json:"venue"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Description string `json:"description"`
	TicketURL   string `json:"ticketUrl"`
}

type CategoryResult struct {
	Events []Event
	Cached bool
	Err    error
}

type CacheClearResponse struct {
	Cleared int `json:"cleared"`
}
