package controllers

import (
	"net/http"

	"event-explorer/utils"
)

type RedirectController struct {
	BaseController
}

// Tickets handles GET /redirect/:eventId. It is not a page: the destination
// always comes from Ticketmaster on the server, never from the visitor, and
// must pass utils.ValidateTicketURL before the 302.
//
// @router /:eventId [get]
func (r *RedirectController) Tickets() {
	eventID, err := utils.ParseEventID(r.Ctx.Input.Param(":eventId"))
	if err != nil {
		r.RenderError(err)
		return
	}

	event, err := eventService.Event(r.Ctx.Request.Context(), eventID)
	if err != nil {
		r.RenderError(err)
		return
	}

	target, err := utils.ValidateTicketURL(event.TicketURL)
	if err != nil {
		r.RenderError(err)
		return
	}

	r.Redirect(target, http.StatusFound)
}
