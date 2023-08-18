package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName     string `gorm:"not null" json:"firstname"`
	LastName      string `gorm:"not null" json:"lastname"`
	Email         string `gorm:"unique, not null" json:"email"`
	Password      string `gorm:"not null" json:"password"`
	BorrowedBooks []Book `json:"borrowedbooks"`
	IsAdmin       bool   `json: "isadmin"`
}
