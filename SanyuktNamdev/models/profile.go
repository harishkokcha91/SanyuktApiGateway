package models

import (
	"time"
)

type Profile struct {
	ID               uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId           uint   `json:"user_id"`
	ProfileFor       string `json:"profileFor" binding:"required,max=50"`
	Name             string `json:"name" binding:"required,min=2,max=100"`
	Image            string `json:"image"`
	DateOfBirth      string `json:"dateOfBirth" binding:"required,datetime=2006-01-02"`
	BirthPlace       string `json:"birthPlace" binding:"max=100"`
	Height           string `json:"height" binding:"max=20"`
	Complexion       string `json:"complexion" binding:"max=50"`
	GotraSelf        string `json:"gotraSelf" binding:"max=100"`
	GotraMother      string `json:"gotraMother" binding:"max=100"`
	GotraGrandMother string `json:"gotraGrandMother" binding:"max=100"`
	Manglik          bool   `json:"manglik"`
	FatherName       string `json:"fatherName" binding:"max=100"`
	FatherOccupation string `json:"fatherOccupation" binding:"max=100"`
	MotherName       string `json:"motherName" binding:"max=100"`
	MotherOccupation string `json:"motherOccupation" binding:"max=100"`
	Siblings         string `json:"siblings" binding:"max=100"`
	Qualification    string `json:"qualification" binding:"max=100"`
	Occupation       string `json:"occupation" binding:"max=100"`
	AnnualIncome     string `json:"annualIncome" binding:"max=50"`
	MaritalStatus    string `json:"maritalStatus" binding:"required,oneof=Single Married Divorced Widowed"`
	Address          string `json:"address" binding:"max=255"`
	CurrentLocation  string `json:"currentLocation" binding:"max=100"`
	Status           string `json:"status" binding:"required,oneof=pending active inactive"`
	PhoneNumbers     string `json:"phoneNumbers" binding:"required,max=20"`

	// Auto-managed timestamps
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}