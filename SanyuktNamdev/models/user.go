package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Password     string
	Email        string `gorm:"unique"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	Image        string `json:"image"`
	PhoneNumbers string `json:"phoneNumbers"`
	Status       string `json:"status"`
	Role         string `gorm:"default:'user'" json:"role"`
}

type UserProfile struct {
	ID            uint      `gorm:"primaryKey"`
	UserID        uint      `gorm:"index" binding:"required" json:"user_id"` // Foreign key for User
	FirstName     string    `gorm:"size:100"`
	LastName      string    `gorm:"size:100"`
	DateOfBirth   time.Time `gorm:"not null"`  // Date of Birth (use time.Time for date handling)
	Age           int       `gorm:"default:0"` //`gorm:"check:age >= 18"` // Enforcing minimum age for matrimony
	Gender        string    `gorm:"size:20"`   // E.g., Male, Female, Other
	FatherName    string    `gorm:"size:100"`  // Father's name
	MotherName    string    `gorm:"size:100"`  // Mother's name
	Education     string    `gorm:"size:100"`  // Education details
	Income        float64   `gorm:"default:0"` // Monthly/annual income
	Religion      string    `gorm:"size:50"`
	MaritalStatus string    `gorm:"size:50"`   // E.g., Single, Divorced, Widowed
	Height        float64   `gorm:"default:0"` //`gorm:"check:height >= 3.0"` // Height in meters
	Weight        float64   `gorm:"default:0"` //`gorm:"check:weight >= 30"` // Weight in kg
	Location      string    `gorm:"size:100"`
	Interests     string    `gorm:"size:255"` // Hobbies or interests
	Language      string    `gorm:"size:100"` // Preferred language
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
	Status        string    `gorm:"size:50"` // Profile status, e.g., Active, Inactive

	// Family Details (Optional)
	FatherOccupation string `gorm:"size:100"`
	MotherOccupation string `gorm:"size:100"`
	SiblingCount     int    `gorm:"default:0"`

	// Gotra (Self, Mother, Grandmother)
	SelfGotra        string `gorm:"size:100"` // Gotra of the individual
	MotherGotra      string `gorm:"size:100"` // Gotra of the mother
	GrandmotherGotra string `gorm:"size:100"` // Gotra of the grandmother

	// Partner Preferences
	PreferredAgeFrom  int `gorm:"default:18"`
	PreferredAgeTo    int `gorm:"default:40"`
	PreferredHeight   float64
	PreferredReligion string `gorm:"size:50"`
	PreferredLocation string `gorm:"size:100"`
	PreferredStatus   string `gorm:"size:50"` // E.g., Single, Widowed, Divorced
}

// Custom Date Format for "YYYY-MM-DD"
const customDateFormat = "2006-01-02"

// UnmarshalJSON custom method to parse dateOfBirth as a string and convert it to time.Time
func (m *UserProfile) UnmarshalJSON(data []byte) error {
	type Alias UserProfile
	aux := &struct {
		DateOfBirth string `json:"dateOfBirth"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	// Unmarshal the JSON into auxiliary structure
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse the DateOfBirth field to time.Time
	date, err := time.Parse(customDateFormat, aux.DateOfBirth)
	if err != nil {
		return fmt.Errorf("invalid date format for DateOfBirth, expected YYYY-MM-DD: %v", err)
	}
	m.DateOfBirth = date

	return nil
}
