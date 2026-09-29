package controllers

import (
	"event-explorer/services"
	"event-explorer/utils"
	"net/http"
)

var locationService *services.LocationService

type LocationController struct {
	BaseController
}

// @Title Autocomplete cities
// @Description Returns up to five city suggestions.
// @Param input query string true "The input text to autocomplete. It must be at least 3 characters long."
// @Param sessionToken query string true "1-36 letters, digits, underscores or hyphens (a UUID fits)"
// @Success 200 {object} models.AutocompleteResponse
// @Success 400 {object} models.ErrorResponse
// @Success 502 {object} models.ErrorResponse
// @router /autocomplete [get]
func (l *LocationController) Autocomplete() {
	params, err := utils.ParseAutocompleteParams(l.Ctx.Request.URL.Query())
	if err != nil {
		l.RespondError(err)
		return
	}

	result, err := locationService.Autocomplete(l.Ctx.Request.Context(), params)
	if err != nil {
		l.RespondError(err)
		return
	}

	l.RespondJSON(http.StatusOK, result)
}

func SetLocationService(s *services.LocationService) {
	locationService = s
}
