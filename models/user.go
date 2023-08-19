package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName     string
	LastName      string
	Email         string `gorm:"unique"`
	Password      string
	Referal       string `gorm:"default:'abc'"`
	IsAdmin       bool   `gorm:"default:false"`
	BorrowedBooks []Book `gorm:"many2many:borrowings;"`
}
