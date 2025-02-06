package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Password     string
	Email        string `gorm:"unique"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	Image        string `json:"image"`
	PhoneNumbers string `json:"phoneNumbers"`
	Status       string `json:"status"`
}
