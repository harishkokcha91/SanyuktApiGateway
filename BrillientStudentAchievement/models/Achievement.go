package models

import (
	"time"
)

type Achievement struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `json:"name" binding:"required"`
	AchievementType   string    `json:"achievement_type" binding:"required"`
	Achievement       string    `json:"achievement" binding:"required"`
	Description       string    `json:"description"`
	DateOfAchievement string    `json:"date_of_achievement"`
	Image             string    `json:"image"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Status            string    `json:"status" gorm:"default:Pending"` // Default status
}
