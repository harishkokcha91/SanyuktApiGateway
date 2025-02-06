package models

import "time"

type Profile struct {
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId           uint      `json:"user_id"`
	ProfileFor       string    `json:"profileFor"`
	Name             string    `json:"name"`
	DateOfBirth      time.Time `json:"date_of_birth"`
	BirthPlace       string    `json:"birth_place"`
	Height           string    `json:"height"`
	Complexion       string    `json:"complexion"`
	GotraSelf        string    `json:"gotra_self"`
	GotraMother      string    `json:"gotra_mother"`
	GotraGrandMother string    `json:"gotra_grandmother"`
	Manglik          bool      `json:"manglik"`
	FatherName       string    `json:"father_name"`
	FatherOccupation string    `json:"father_occupation"`
	MotherName       string    `json:"mother_name"`
	MotherOccupation string    `json:"mother_occupation"`
	Siblings         string    `json:"siblings"` // Can store details as JSON or CSV
	Qualification    string    `json:"qualification"`
	Occupation       string    `json:"occupation"`
	AnnualIncome     string    `json:"annual_income"`
	MaritalStatus    string    `json:"marital_status"`
	Address          string    `json:"address"`
	CurrentLocation  string    `json:"current_location"`
	PhoneNumbers     string    `json:"phone_numbers"` // Can store multiple numbers as JSON or CSV
}
