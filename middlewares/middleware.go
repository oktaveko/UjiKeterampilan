// middleware/middleware.go

package middlewares

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strings"
	"ujiketerampilan/models" // Ganti dengan path yang sesuai
)

func JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			tokenString := ctx.Request().Header.Get("Authorization")
			if tokenString == "" {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Missing token"})
			}

			tokenParts := strings.Split(tokenString, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token format"})
			}

			token, err := jwt.Parse(tokenParts[1], func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWTSK")), nil // Ganti dengan secret key Anda
			})
			if err != nil || !token.Valid {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token claims"})
			}

			userID := uint(claims["user_id"].(float64))
			ctx.Set("user_id", userID)

			return next(ctx)
		}
	}
}

func IsAdminMiddleware(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			userID := ctx.Get("user_id").(uint)

			var user models.User
			// Mengasumsikan db adalah koneksi database GORM yang valid
			if err := db.First(&user, userID).Error; err != nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "User not found"})
			}

			if !user.IsAdmin {
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized middle"})
			}

			return next(ctx)
		}
	}
}

func AdminAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			isAdminAuthEnabled := ctx.Get("ISADMINAUTH").(string)
			postedAdminAuth := ctx.Request().FormValue("admin_auth")

			if isAdminAuthEnabled != "" && postedAdminAuth != isAdminAuthEnabled {
				return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Unauthorized"})
			}

			return next(ctx)
		}
	}
}
