package models

import "gorm.io/gorm"

type Business struct {
	gorm.Model
	Name           string   `json:"name" binding:"required"`
	Category       string   `json:"category" binding:"required"`
	Description    string   `json:"description"`
	Owner          string   `json:"owner"` // Added Owner
	Email          string   `json:"email"` // Moved up for better readability
	Phone          string   `json:"phone"`
	WhatsApp       string   `json:"whatsapp"`
	Location       string   `json:"location"` // Added Location
	Address        string   `json:"address"`
	City           string   `json:"city"`     // Added City
	State          string   `json:"state"`    // Added State
	ZipCode        string   `json:"zip_code"` // Added Zip Code
	Country        string   `json:"country"`  // Added Country
	Website        string   `json:"website"`
	Image          string   `json:"image"`
	Status         string   `json:"status" gorm:"default:'Active'"`   // Default to "Active"
	IsVerified     bool     `json:"is_verified" gorm:"default:false"` // Default to false
	OpeningHours   string   `json:"opening_hours"`
	HomeDelivery   bool     `json:"home_delivery"`
	PaymentMethods []string `gorm:"type:text" json:"payment_methods"`
}
