package main

import (
	"os"

	"event-explorer/controllers"
	_ "event-explorer/routers"
	"event-explorer/services"
	"event-explorer/services/googleplacesapi"
	"event-explorer/services/ticketmasterapi"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logs.Info("No .env file found, using system environment variables")
	}

	googleKey := requireEnv("GOOGLE_PLACES_API_KEY")
	ticketmasterKey := requireEnv("TICKETMASTER_API_KEY")

	controllers.SetLocationService(services.NewLocationService(googleplacesapi.NewClient(googleKey)))
	controllers.SetEventService(services.NewEventService(ticketmasterapi.NewClient(ticketmasterKey)))

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}

func requireEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		logs.Critical("%s is not set", name)
		os.Exit(1)
	}

	return value
}
