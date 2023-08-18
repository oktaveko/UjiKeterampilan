package main

import (
	"UjiKeterampilan/configs"
	"UjiKeterampilan/routes"
	"github.com/labstack/echo/v4"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// loadEnv()
	configs.InitDatabase()
	e := echo.New()
	routes.InitRoute(e)

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
