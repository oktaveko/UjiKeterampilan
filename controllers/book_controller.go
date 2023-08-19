package controllers

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"ujiketerampilan/models"
)

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
