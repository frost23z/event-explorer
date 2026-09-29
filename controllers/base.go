package controllers

import (
	"context"
	"errors"
	"event-explorer/models"
	"net/http"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
)

const statusClientClosedRequest = 499

type BaseController struct {
	beego.Controller
}

func (b *BaseController) RespondJSON(status int, v any) {
	b.Ctx.Output.SetStatus(status)
	b.Data["json"] = v

	if err := b.ServeJSON(); err != nil {
		logs.Error(
			"%s %s failed to serve JSON response: %v",
			b.Ctx.Request.Method,
			b.Ctx.Request.URL.Path,
			err,
		)
	}
}

func (b *BaseController) RespondError(err error) {
	if errors.Is(err, context.Canceled) {
		logs.Debug("%s %s canceled by client", b.Ctx.Request.Method, b.Ctx.Request.URL.Path)
		b.Ctx.Output.SetStatus(statusClientClosedRequest)
		return
	}

	apiErr := models.AsAPIError(err)

	if apiErr.StatusCode >= http.StatusInternalServerError {
		logs.Error(
			"%s %s failed with %d: %v",
			b.Ctx.Request.Method,
			b.Ctx.Request.URL.Path,
			apiErr.StatusCode,
			apiErr.Err,
		)
	}

	b.RespondJSON(
		apiErr.StatusCode,
		models.ErrorResponse{Error: apiErr.Message},
	)
}
