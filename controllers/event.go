package controllers

import (
	"net/http"

	"event-explorer/services"
	"event-explorer/utils"
)

var eventService *services.EventService

type EventController struct {
	BaseController
}

// @Title List events
// @Description Returns up to six events of one category in a city.
// @Param city query string true "City name returned by /api/locations/:placeId"
// @Param countryCode query string true "Two-letter country code returned by /api/locations/:placeId"
// @Param category query string true "Music or Sports"
// @Success 200 {array} models.Event
// @Success 400 {object} models.ErrorResponse
// @Success 502 {object} models.ErrorResponse
// @router / [get]
func (e *EventController) List() {
	params, err := utils.ParseEventSearchParams(e.Ctx.Request.URL.Query())
	if err != nil {
		e.RespondError(err)
		return
	}

	result, err := eventService.Events(e.Ctx.Request.Context(), params)
	if err != nil {
		e.RespondError(err)
		return
	}

	e.RespondJSON(http.StatusOK, result)
}

// @Title Event details
// @Description Returns one event, including its ticket URL.
// @Param eventId path string true "Event id returned by /api/events"
// @Success 200 {object} models.Event
// @Success 400 {object} models.ErrorResponse
// @Success 404 {object} models.ErrorResponse
// @Success 502 {object} models.ErrorResponse
// @router /:eventId [get]
func (e *EventController) Details() {
	eventID, err := utils.ParseEventID(e.Ctx.Input.Param(":eventId"))
	if err != nil {
		e.RespondError(err)
		return
	}

	result, err := eventService.Event(e.Ctx.Request.Context(), eventID)
	if err != nil {
		e.RespondError(err)
		return
	}

	e.RespondJSON(http.StatusOK, result)
}

func SetEventService(s *services.EventService) {
	eventService = s
}
