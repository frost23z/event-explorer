package controllers

import (
	"event-explorer/utils"

	"github.com/beego/beego/v2/core/logs"
)

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

	logs.Info("%v", params)
}
