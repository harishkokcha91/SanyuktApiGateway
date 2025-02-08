package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Password     string `json:"-"` // Hide password in JSON responses
	Email        string `gorm:"unique" json:"email"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	Image        string `json:"image"`
	PhoneNumbers string `json:"phoneNumbers"`
	Status       string `gorm:"default:'pending'" json:"status"`
}
