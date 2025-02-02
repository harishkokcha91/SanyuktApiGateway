package models

import "time"

type UserDetails struct {
	UserId    int       `json:"user_id" gorm:"unique;not null"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
