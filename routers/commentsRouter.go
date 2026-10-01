package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

    beego.GlobalControllerRouter["event-explorer/controllers:CacheController"] = append(beego.GlobalControllerRouter["event-explorer/controllers:CacheController"],
        beego.ControllerComments{
            Method: "Clear",
            Router: `/`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["event-explorer/controllers:EventController"] = append(beego.GlobalControllerRouter["event-explorer/controllers:EventController"],
        beego.ControllerComments{
            Method: "Listing",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["event-explorer/controllers:EventController"] = append(beego.GlobalControllerRouter["event-explorer/controllers:EventController"],
        beego.ControllerComments{
            Method: "Details",
            Router: `/:eventId`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["event-explorer/controllers:LocationController"] = append(beego.GlobalControllerRouter["event-explorer/controllers:LocationController"],
        beego.ControllerComments{
            Method: "Details",
            Router: `/:placeId`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["event-explorer/controllers:LocationController"] = append(beego.GlobalControllerRouter["event-explorer/controllers:LocationController"],
        beego.ControllerComments{
            Method: "Autocomplete",
            Router: `/autocomplete`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["event-explorer/controllers:RedirectController"] = append(beego.GlobalControllerRouter["event-explorer/controllers:RedirectController"],
        beego.ControllerComments{
            Method: "Tickets",
            Router: `/:eventId`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
