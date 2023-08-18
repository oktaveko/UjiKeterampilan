package routes

import (
	"UjiKeterampilan/controllers"
	"github.com/labstack/echo/v4"
	"os"

	echojwt "github.com/labstack/echo-jwt"
	"github.com/labstack/echo/v4/middleware"
)

func InitRoute(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.POST("/login", controllers.LoginController)

	auth := e.Group("")
	auth.Use(echojwt.JWT([]byte(os.Getenv("SECRET_KEY_JWT"))))
	auth.GET("/books", controllers.GetBooksController)
	auth.GET("/users", controllers.GetUsersController)
	auth.GET("/books/:id", controllers.GetDetailBookController)
	auth.POST("/books", controllers.AddBookController)
}
