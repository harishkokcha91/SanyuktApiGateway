package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Password     string `json:"-" binding:"required,min=8,max=72"`
	Email        string `json:"email" gorm:"unique" binding:"required,email"`
	Name         string `json:"name" binding:"required,min=2,max=100"`
	Age          int    `json:"age" binding:"min=1,max=120"`
	Image        string `json:"image"`
	PhoneNumbers string `json:"phoneNumbers" binding:"max=20"`
	Status       string `json:"status" binding:"oneof=active inactive pending"`
	Role         string `gorm:"default:'user'" json:"role" binding:"oneof=user admin"`
}
