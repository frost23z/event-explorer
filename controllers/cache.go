package controllers

import (
	"net/http"

	"event-explorer/models"
)

type CacheController struct {
	BaseController
}

// Clear handles DELETE /api/cache. It empties the event cache, so the next
// search asks Ticketmaster again, and reports how many entries it removed.
//
// @router / [delete]
func (c *CacheController) Clear() {
	c.RespondJSON(http.StatusOK, models.CacheClearResponse{Cleared: eventService.ClearCache()})
}
