package models

import (
	"time"
)

type Achievement struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `json:"name" binding:"required,min=2,max=200"`
	AchievementType   string    `json:"achievement_type" binding:"required,max=100"`
	Achievement       string    `json:"achievement" binding:"required,max=2000"`
	Description       string    `json:"description" binding:"max=2000"`
	DateOfAchievement string    `json:"date_of_achievement" binding:"required,datetime=2006-01-02"`
	Image             string    `json:"image"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Status            string    `json:"status" gorm:"default:Pending" binding:"omitempty,oneof=Pending Approved Rejected"`

	// Approval audit fields
	ApprovedBy  *uint      `json:"approved_by" gorm:"index"`
	ApprovedAt  *time.Time `json:"approved_at"`
	RejectedBy  *uint      `json:"rejected_by" gorm:"index"`
	RejectedAt  *time.Time `json:"rejected_at"`
}