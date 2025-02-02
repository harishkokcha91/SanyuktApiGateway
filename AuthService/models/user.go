package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Password string
	Email    string `gorm:"unique"`
	Name     string `json:"name"`
}
