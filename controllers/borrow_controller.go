package controllers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
	"ujiketerampilan/models"
)

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

		// Menghapus semua BorrowedBooks yang sudah dikembalikan
		if err := tx.Where("user_id = ? AND returned = ?", ctx.Get("user_id").(uint), true).Unscoped().Delete(&models.Borrowing{}).Error; err != nil {
			return err
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

func (c *Controller) ManualReturn(ctx echo.Context) error {
	userID, _ := strconv.Atoi(ctx.Param("user_id"))
	authenticatedUserID := ctx.Get("user_id").(uint) // Ambil user_id dari JWT token

	if userID != int(authenticatedUserID) {
		return ctx.JSON(http.StatusForbidden, map[string]string{"error": "Forbidden: You can return books for your own user_id"})
	}

	var user models.User
	if err := c.db.Preload("BorrowedBooks").First(&user, userID).Error; err != nil {
		// Periksa apakah error adalah record not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"message": "You don't have any borrowed books to return"})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get user information"})
	}

	// Menggunakan transaksi untuk mengembalikan semua buku yang dipinjam
	err := c.db.Transaction(func(tx *gorm.DB) error {
		var bookIDs []uint
		for _, book := range user.BorrowedBooks {
			bookIDs = append(bookIDs, book.ID)
		}

		if err := c.ReturnBook(ctx, tx, bookIDs); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to return books"})
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "Books returned successfully"})

}
