package main

import (
	"os"

	"event-explorer/controllers"
	_ "event-explorer/routers"
	"event-explorer/services"
	"event-explorer/services/googleplacesapi"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logs.Info("No .env file found, using system environment variables")
	}

	if apiKey := os.Getenv("GOOGLE_PLACES_API_KEY"); apiKey == "" {
		logs.Critical("GOOGLE_PLACES_API_KEY is not set")
		os.Exit(1)
	} else {
		controllers.SetLocationService(services.NewLocationService(googleplacesapi.NewClient(apiKey)))
	}

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}
