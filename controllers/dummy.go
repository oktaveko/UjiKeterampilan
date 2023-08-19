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
	if user.Referal == "1234" {
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

func (c *Controller) GetBooks(ctx echo.Context) error {
	var books []models.Book
	c.db.Select("title", "author", "available_quantity").Find(&books)

	// Membuat slice baru untuk menyimpan informasi yang dipilih
	var simplifiedBooks []map[string]interface{}
	for _, book := range books {
		simplifiedBook := map[string]interface{}{
			"title":         book.Title,
			"author":        book.Author,
			"unitAvailable": book.AvailableQuantity,
		}
		simplifiedBooks = append(simplifiedBooks, simplifiedBook)
	}

	return ctx.JSON(http.StatusOK, simplifiedBooks)
}

func (c *Controller) GetBookByID(ctx echo.Context) error {
	bookID, _ := strconv.Atoi(ctx.Param("book_id"))

	var book models.Book
	if err := c.db.First(&book, bookID).Error; err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	return ctx.JSON(http.StatusOK, book)
}

func (c *Controller) BorrowBooks(ctx echo.Context) error {
	requestedUserID, _ := strconv.Atoi(ctx.Param("user_id"))
	authenticatedUserID := ctx.Get("user_id").(uint) // Ambil user_id dari JWT token

	if requestedUserID != int(authenticatedUserID) {
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: You can only borrow books for your own user_id"})
	}

	var user models.User
	if err := c.db.Preload("BorrowedBooks").First(&user, requestedUserID).Error; err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	var borrowingData struct {
		BookIDs []uint `json:"book_ids"`
	}
	if err := ctx.Bind(&borrowingData); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data"})
	}

	// Memeriksa apakah book_ids tidak kosong
	if len(borrowingData.BookIDs) == 0 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Book IDs cannot be empty"})
	}

	// Memeriksa jumlah buku yang sedang dipinjam oleh pengguna
	if len(user.BorrowedBooks)+len(borrowingData.BookIDs) > 3 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Exceeded borrowing limit"})
	}

	for _, bookID := range borrowingData.BookIDs {
		var book models.Book
		if err := c.db.First(&book, bookID).Error; err != nil {
			return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
		}

		if book.AvailableQuantity == 0 {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Book not available"})
		}

		borrowing := models.Borrowing{UserID: user.ID, BookID: book.ID, DueDate: time.Now().AddDate(0, 0, 7), Returned: false}
		c.db.Create(&borrowing)

		book.AvailableQuantity--
		c.db.Save(&book)
	}

	return ctx.JSON(http.StatusCreated, map[string]string{"message": "Books borrowed successfully"})
}

func (c *Controller) GetUserByID(ctx echo.Context) error {
	userID, _ := strconv.Atoi(ctx.Param("user_id"))

	var user models.User
	if err := c.db.First(&user, userID).Error; err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"FirstName":     user.FirstName,
		"LastName":      user.LastName,
		"BorrowedBooks": user.BorrowedBooks,
	})
}

func (c *Controller) ReturnBook(ctx echo.Context, tx *gorm.DB, bookIDs []uint) error {
	return tx.Transaction(func(tx *gorm.DB) error {
		for _, bookID := range bookIDs {
			var borrowing models.Borrowing
			if err := tx.Where("book_id = ? AND returned = ?", bookID, false).First(&borrowing).Error; err != nil {
				continue // Lanjutkan jika tidak ada peminjaman yang sesuai
			}

			var book models.Book
			if err := tx.First(&book, bookID).Error; err != nil {
				return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
			}

			book.AvailableQuantity++
			if err := tx.Save(&book).Error; err != nil {
				return err // Rollback akan terjadi karena error saat melakukan perubahan pada buku
			}

			borrowing.Returned = true
			if err := tx.Save(&borrowing).Error; err != nil {
				return err // Rollback akan terjadi karena error saat melakukan perubahan pada peminjaman
			}
		}
		return nil
	})
}

func (c *Controller) AutomateReturnBooks() {
	tx := c.db.Begin()

	var overdueBorrowings []models.Borrowing
	if err := tx.Where("due_date <= ? AND returned = ?", time.Now(), false).Find(&overdueBorrowings).Error; err != nil {
		tx.Rollback()
		return
	}

	for _, borrowing := range overdueBorrowings {
		borrowing.Returned = true

		var book models.Book
		if err := tx.First(&book, borrowing.BookID).Error; err != nil {
			tx.Rollback()
			return
		}
		book.AvailableQuantity++
		if err := tx.Save(&book).Error; err != nil {
			tx.Rollback()
			return
		}

		if err := tx.Save(&borrowing).Error; err != nil {
			tx.Rollback()
			return
		}
	}

	tx.Commit()
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

func (c *Controller) CreateBook(ctx echo.Context) error {
	var book models.Book
	if err := ctx.Bind(&book); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data"})
	}

	// Memastikan semua field yang diperlukan terisi
	if book.Title == "" || book.Author == "" || book.PublishedAt.IsZero() || book.AvailableQuantity == 0 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
	}

	c.db.Create(&book)

	return ctx.JSON(http.StatusCreated, book)
}

func (c *Controller) UpdateBook(ctx echo.Context) error {
	bookID, _ := strconv.Atoi(ctx.Param("book_id"))

	var book models.Book
	if err := c.db.First(&book, bookID).Error; err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Book not found"})
	}

	var updatedBook models.Book
	if err := ctx.Bind(&updatedBook); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid data"})
	}

	c.db.Model(&book).Updates(updatedBook)

	return ctx.JSON(http.StatusOK, book)
}
