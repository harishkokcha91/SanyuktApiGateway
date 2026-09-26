package models

import (
	"time"

	"gorm.io/gorm"
)

type Business struct {
	gorm.Model
	Name           string   `json:"name" binding:"required,min=2,max=200"`
	Category       string   `json:"category" binding:"required,max=100"`
	Description    string   `json:"description" binding:"max=2000"`
	Owner          string   `json:"owner" binding:"required,max=100"`
	Email          string   `json:"email" binding:"required,email"`
	Phone          string   `json:"phone" binding:"max=20"`
	WhatsApp       string   `json:"whatsapp" binding:"max=20"`
	Location       string   `json:"location" binding:"required,max=200"`
	Address        string   `json:"address" binding:"max=300"`
	City           string   `json:"city" binding:"required,max=100"`
	State          string   `json:"state" binding:"required,max=100"`
	ZipCode        string   `json:"zip_code" binding:"max=20"`
	Country        string   `json:"country" binding:"required,max=100"`
	Website        string   `json:"website" binding:"url"`
	Image          string   `json:"image"`
	Status         string   `json:"status" gorm:"default:'Pending'" binding:"required,oneof=Pending Approved Rejected"`
	IsVerified     bool     `json:"is_verified" gorm:"default:false"`
	OpeningHours   string   `json:"opening_hours" binding:"max=100"`
	HomeDelivery   bool     `json:"home_delivery"`
	PaymentMethods []string `gorm:"type:text" json:"payment_methods"`

	// Approval audit fields
	ApprovedBy  *uint      `json:"approved_by" gorm:"index"`
	ApprovedAt  *time.Time `json:"approved_at"`
	RejectedBy  *uint      `json:"rejected_by" gorm:"index"`
	RejectedAt  *time.Time `json:"rejected_at"`
}