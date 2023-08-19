package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"os"
	"time"
	"ujiketerampilan/configs"
	"ujiketerampilan/controllers"
	"ujiketerampilan/routes"
)

func main() {
	// loadEnv()
	e := echo.New()

	configs.InitDatabase()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Inisialisasi controller dengan database
	controller := controllers.NewController(configs.DB)

	// Routing
	routes.SetupRoutes(e, controller)

	go func() {
		for {
			controller.AutomateReturnBooks()
			time.Sleep(1 * time.Hour)
		}
	}()

	e.Start(getPort())
}

func getPort() string {
	if envPort := os.Getenv("PORT"); envPort != "" {
		return ":" + envPort
	}
	return ":8080"
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}
