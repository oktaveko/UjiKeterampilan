package models

import (
	"gorm.io/gorm"
	"time"
)

type Borrowing struct {
	gorm.Model
	UserID   uint
	BookID   uint
	DueDate  time.Time
	Returned bool
}
