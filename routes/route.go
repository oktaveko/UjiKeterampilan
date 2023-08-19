package routes

import (
	"github.com/labstack/echo/v4"
	"os"
	"ujiketerampilan/configs"
	"ujiketerampilan/controllers"
	"ujiketerampilan/middlewares"

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

func SetupRoutes(e *echo.Echo, c *controllers.Controller) {
	// Public Routes
	e.POST("/register", c.RegisterUser)
	e.POST("/login", c.LoginUser)
	e.GET("/books", c.GetBooks)
	e.GET("/books/:book_id", c.GetBookByID)
	e.GET("/users/:user_id", c.GetUserByID)

	// Protected Routes
	protected := e.Group("/myaccount")
	protected.Use(middlewares.JWTMiddleware())
	protected.POST("/users/:user_id/borrow", c.BorrowBooks)
	protected.DELETE("/users/:user_id/delete", c.DeleteUser)

	// Admin-Only Routes
	admin := e.Group("/admin")
	admin.Use(middlewares.JWTMiddleware(), middlewares.IsAdminMiddleware(configs.DB))
	//admin.PUT("/verif", c.SetUserAsAdmin)
	admin.POST("/books", c.CreateBook)
	admin.PUT("/books/:book_id", c.UpdateBook)
}
