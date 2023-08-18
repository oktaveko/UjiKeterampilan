package models

import (
	"gorm.io/gorm"
	"time"
)

type Book struct {
	gorm.Model
	Title             string `gorm:"not null" json:"title"`
	Author            string `gorm:"not null" json:"author"`
	PublishedAt       time.Time
	AvailableQuantity int
}
