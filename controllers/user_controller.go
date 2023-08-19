package controllers

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"time"
	"ujiketerampilan/models"
)

type Controller struct {
	db *gorm.DB
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{db: db}
}

func GenerateToken(userID uint) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString([]byte(os.Getenv("JWTSK"))) // Ganti dengan secret key Anda
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (c *Controller) RegisterUser(ctx echo.Context) error {
	var user models.User
	if err := ctx.Bind(&user); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data"})
	}

	// Cek apakah email sudah terdaftar
	var existingUser models.User
	if err := c.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return ctx.JSON(http.StatusConflict, map[string]string{"error": "Email is already registered"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	user.Password = string(hashedPassword)

	// Cek Referal dan set IsAdmin jika cocok
	if user.Referal == os.Getenv("REFERAL_KEY") {
		user.IsAdmin = true
	}

	c.db.Create(&user)

	token, err := GenerateToken(user.ID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return ctx.JSON(http.StatusCreated, map[string]string{"token": token})
}

func (c *Controller) LoginUser(ctx echo.Context) error {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := ctx.Bind(&loginData); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data"})
	}

	var user models.User
	if err := c.db.Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)); err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"token": token})
}

func (c *Controller) GetUserByID(ctx echo.Context) error {
	userID, _ := strconv.Atoi(ctx.Param("user_id"))

	var user models.User
	if err := c.db.First(&user, userID).Error; err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	var borrowings []models.Borrowing
	if err := c.db.Model(&models.Borrowing{}).Where("user_id = ?", userID).Pluck("book_id", &borrowings).Error; err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get borrowed books"})
	}

	var borrowedBookIDs []uint
	for _, borrowing := range borrowings {
		borrowedBookIDs = append(borrowedBookIDs, borrowing.BookID)
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"FirstName":     user.FirstName,
		"LastName":      user.LastName,
		"BorrowedBooks": borrowedBookIDs,
	})
}

func (c *Controller) DeleteUser(ctx echo.Context) error {
	userID, _ := strconv.Atoi(ctx.Param("user_id"))
	authenticatedUserID := ctx.Get("user_id").(uint) // Ambil user_id dari JWT token

	if userID != int(authenticatedUserID) {
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: You can only delete account for your own user_id"})
	}

	var user models.User
	if err := c.db.Preload("BorrowedBooks").First(&user, userID).Error; err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	var passwordData struct {
		Password string `json:"password"`
	}
	if err := ctx.Bind(&passwordData); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordData.Password)); err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid password"})
	}

	// Menggunakan transaksi untuk menghapus pengguna dan mengembalikan semua buku yang dipinjam
	err := c.db.Transaction(func(tx *gorm.DB) error {
		var bookIDs []uint
		for _, book := range user.BorrowedBooks {
			bookIDs = append(bookIDs, book.ID)
		}
		if err := c.ReturnBook(ctx, tx, bookIDs); err != nil {
			return err
		}

		// Menghapus pengguna
		if err := tx.Unscoped().Delete(&user).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete user"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "User deleted successfully"})
}
