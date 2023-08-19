package models

import (
	"gorm.io/gorm"
	"time"
)

type Book struct {
	gorm.Model
	Title             string
	Author            string
	PublishedAt       time.Time
	AvailableQuantity uint
	Borrowings        []Borrowing `gorm:"many2many:borrowings;"`
}
