package controllers

import (
	"UjiKeterampilan/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"time"
)

type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func LoginController(c echo.Context) error {
	var loginRequest models.User
	c.Bind(&loginRequest)

	// bisnis logic databse login

	claims := &jwtCustomClaims{
		"Alterra Academy",
		true,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(os.Getenv("SECRET_KEY_JWT")))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.BaseResponse{
		Status:  true,
		Message: "Berhasil",
		Data: map[string]string{
			"token": t,
		}})
}
