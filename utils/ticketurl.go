package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"event-explorer/models"
)

// approvedTicketHosts are the only hosts /redirect/:eventId may send visitors
// to. A host matches when it is one of these or a subdomain of one.
var approvedTicketHosts = []string{
	// Ticketmaster country domains
	"ticketmaster.com",
	"ticketmaster.ca",
	"ticketmaster.co.uk",
	"ticketmaster.ie",
	"ticketmaster.com.au",
	"ticketmaster.co.nz",
	"ticketmaster.com.mx",
	"ticketmaster.com.br",
	"ticketmaster.cl",
	"ticketmaster.co.za",
	"ticketmaster.ae",
	"ticketmaster.de",
	"ticketmaster.fr",
	"ticketmaster.es",
	"ticketmaster.it",
	"ticketmaster.nl",
	"ticketmaster.be",
	"ticketmaster.at",
	"ticketmaster.ch",
	"ticketmaster.dk",
	"ticketmaster.fi",
	"ticketmaster.no",
	"ticketmaster.se",
	"ticketmaster.pl",
	"ticketmaster.cz",

	// Affiliate redirect the Discovery API returns for some events
	"ticketmaster.evyy.net",

	// Other ticket sellers
	"livenation.com",
	"ticketweb.com",
	"universe.com",
	"frontgatetickets.com",
}

// ValidateTicketURL returns the URL to redirect to, or an error when the URL
// is missing, not an absolute https URL, or not on an approved host.
func ValidateTicketURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return "", ticketUnavailable(fmt.Errorf("ticket url %q is not an absolute https url", raw))
	}

	host := strings.ToLower(u.Hostname())

	for _, approved := range approvedTicketHosts {
		// The "." keeps evilticketmaster.com from matching ticketmaster.com.
		if host == approved || strings.HasSuffix(host, "."+approved) {
			return u.String(), nil
		}
	}

	return "", ticketUnavailable(fmt.Errorf("ticket url host %q is not approved", host))
}

func ticketUnavailable(err error) models.APIError {
	return models.APIError{
		StatusCode: http.StatusBadRequest,
		Message:    models.MsgTicketUnavailable,
		Err:        err,
	}
}
