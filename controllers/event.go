package controllers

import (
	"net/url"
	"strings"
	"time"

	"event-explorer/models"
	"event-explorer/services"
	"event-explorer/utils"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	funcs := map[string]any{"eventDay": eventDay, "eventTime": eventTime, "eventPlace": eventPlace}

	for name, fn := range funcs {
		if err := beego.AddFuncMap(name, fn); err != nil {
			logs.Critical("register template func %s: %v", name, err)
		}
	}
}

var eventService *services.EventService

type EventController struct {
	BaseController
}

type section struct {
	Title   string
	Kind    string // models.CategoryMusic or models.CategorySports; picks the placeholder image
	Tagline string
	Events  []models.Event
	Error   string
}

// Prepare wraps every page (and the error page) in views/layout.tpl.
func (p *EventController) Prepare() {
	p.Layout = "layout.tpl"
}

// @router / [get]
func (p *EventController) Listing() {
	query := p.Ctx.Request.URL.Query()

	params, err := utils.ParseListingParams(query)
	if err != nil {
		p.RenderError(err)
		return
	}

	// params.City is lowercase (it is the cache key); show the city as it was in the URL.
	city, link, _ := listingLink(query)

	music, sports := eventService.MusicAndSports(p.Ctx.Request.Context(), params.City, params.CountryCode)

	// Lets anyone check cache reuse without reading the logs.
	p.Ctx.Output.Header("X-Music-Cache", cacheStatus(music))
	p.Ctx.Output.Header("X-Sports-Cache", cacheStatus(sports))

	p.Data["Title"] = "Events in " + city
	p.Data["City"] = city
	p.Data["CountryCode"] = params.CountryCode
	p.Data["ListingURL"] = link
	p.Data["Sections"] = []section{
		newSection("Music", models.CategoryMusic, "Turn up the evening", music),
		newSection("Sports", models.CategorySports, "Get into the game", sports),
	}
	p.TplName = "listing.tpl"
}

// @router /:eventId [get]
func (p *EventController) Details() {
	eventID, err := utils.ParseEventID(p.Ctx.Input.Param(":eventId"))
	if err != nil {
		p.RenderError(err)
		return
	}

	event, err := eventService.Event(p.Ctx.Request.Context(), eventID)
	if err != nil {
		p.RenderError(err)
		return
	}

	// Back goes to the listing the visitor came from, or to the home page
	// for a direct link with no (valid) city.
	city, link, fromListing := listingLink(p.Ctx.Request.URL.Query())

	p.Data["Title"] = event.Name
	p.Data["Event"] = event
	p.Data["City"] = city
	p.Data["BackURL"] = link
	p.Data["FromListing"] = fromListing
	p.TplName = "details.tpl"
}

func SetEventService(s *services.EventService) {
	eventService = s
}

// cacheStatus is HIT when the section came from the cache. A failed section
// never did, so it is a MISS.
func cacheStatus(result models.CategoryResult) string {
	if result.Cached {
		return "HIT"
	}

	return "MISS"
}

func newSection(title, kind, tagline string, result models.CategoryResult) section {
	if result.Err != nil {
		apiErr := models.AsAPIError(result.Err)
		logs.Error("%s events failed: %v", title, apiErr.Err)

		return section{Title: title, Kind: kind, Tagline: tagline, Error: apiErr.Message}
	}

	return section{Title: title, Kind: kind, Tagline: tagline, Events: result.Events}
}

func listingLink(query url.Values) (city, link string, ok bool) {
	params, err := utils.ParseListingParams(query)
	if err != nil {
		return "", "/", false
	}

	city = strings.TrimSpace(query.Get("city"))
	link = "/events?" + url.Values{"city": {city}, "countryCode": {params.CountryCode}}.Encode()

	return city, link, true
}

// eventDay turns "2026-10-12" into "Mon, 12 Oct 2026". A value that cannot be
// parsed is shown as it is.
func eventDay(date string) string {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}

	return day.Format("Mon, 02 Jan 2006")
}

// eventTime turns "19:30:00" into "7:30 PM", or "" when there is no usable time.
func eventTime(clock string) string {
	at, err := time.Parse("15:04:05", clock)
	if err != nil {
		return ""
	}

	return at.Format("3:04 PM")
}

// eventPlace joins the address and city that are available.
func eventPlace(event models.Event) string {
	var parts []string

	for _, part := range []string{event.Address, event.City} {
		if part != "" {
			parts = append(parts, part)
		}
	}

	return strings.Join(parts, ", ")
}
