package routers

import (
	"event-explorer/controllers"

	beego "github.com/beego/beego/v2/server/web"
)

func init() {
	beego.Router("/", &controllers.HomeController{})

	beego.AddNamespace(
		beego.NewNamespace("/api",
			beego.NSNamespace("/locations",
				beego.NSInclude(&controllers.LocationController{}),
			),
			beego.NSNamespace("/cache",
				beego.NSInclude(&controllers.CacheController{}),
			),
		),
		beego.NewNamespace("/events",
			beego.NSInclude(&controllers.EventController{}),
		),
		beego.NewNamespace("/redirect",
			beego.NSInclude(&controllers.RedirectController{}),
		),
	)
}
