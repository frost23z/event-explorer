package controllers

import (
	"event-explorer/utils"

	"github.com/beego/beego/v2/core/logs"
)

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

	logs.Info("%v", params)
}
