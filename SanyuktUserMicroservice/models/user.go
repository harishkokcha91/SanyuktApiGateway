package models

import "time"

type UserDetails struct {
	UserId       int       `json:"user_id" gorm:"unique;not null"`
	Name         string    `json:"name"`
	Age          int       `json:"age"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Image        string    `json:"image"`
	PhoneNumbers string    `json:"phone_number"`
	Status       string    `json:"status"`
}
