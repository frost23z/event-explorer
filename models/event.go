package models

const (
	CategoryMusic  = "music"
	CategorySports = "sports"

	MaxCityBytes = 100

	MsgInvalidEventSearch = "Use a city, a two-letter country code and either Music or Sports."
	MsgInvalidEvent       = "Use a valid event."
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
