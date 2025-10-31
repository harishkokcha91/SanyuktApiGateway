package models

import (
	"time"
)

type Profile struct {
	ID               uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId           uint   `json:"user_id"`
	ProfileFor       string `json:"profileFor"`
	Name             string `json:"name"`
	Image            string `json:"image"`
	DateOfBirth      string `json:"dateOfBirth"` // Keeping it as string for JSON compatibility
	BirthPlace       string `json:"birthPlace"`
	Height           string `json:"height"`
	Complexion       string `json:"complexion"`
	GotraSelf        string `json:"gotraSelf"`
	GotraMother      string `json:"gotraMother"`
	GotraGrandMother string `json:"gotraGrandMother"`
	Manglik          bool   `json:"manglik"`
	FatherName       string `json:"fatherName"`
	FatherOccupation string `json:"fatherOccupation"`
	MotherName       string `json:"motherName"`
	MotherOccupation string `json:"motherOccupation"`
	Siblings         string `json:"siblings"`
	Qualification    string `json:"qualification"`
	Occupation       string `json:"occupation"`
	AnnualIncome     string `json:"annualIncome"`
	MaritalStatus    string `json:"maritalStatus"`
	Address          string `json:"address"`
	CurrentLocation  string `json:"currentLocation"`
	Status           string `json:"status"`
	PhoneNumbers     string `json:"phoneNumbers"`

	// Auto-managed timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}
