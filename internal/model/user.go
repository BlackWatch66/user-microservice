package model

import (
	"time"

	"gorm.io/gorm"
)

// User model
type User struct {
	gorm.Model
	Email        string `gorm:"type:varchar(100);index;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	FirstName    string `gorm:"type:varchar(50)"`
	LastName     string `gorm:"type:varchar(50)"`
	Addresses    []Address `gorm:"foreignKey:UserID"` // one-to-many relationship
}

// Address model
type Address struct {
	gorm.Model
	UserID      uint   `gorm:"index;not null"` // foreign key
	Street      string `gorm:"type:varchar(255);not null"`
	City        string `gorm:"type:varchar(100);not null"`
	State       string `gorm:"type:varchar(100)"`
	PostalCode  string `gorm:"type:varchar(20);not null"`
	Country     string `gorm:"type:varchar(100);not null"`
	IsDefault   bool   `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
} 