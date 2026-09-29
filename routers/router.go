package routers

import (
	"event-explorer/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	ns := beego.NewNamespace("/api",
		beego.NSNamespace("/locations",
			beego.NSInclude(
				&controllers.LocationController{},
			),
		),
		beego.NSNamespace("/events",
			beego.NSInclude(
				&controllers.EventController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
