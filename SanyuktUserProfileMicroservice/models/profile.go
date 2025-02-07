package models

type Profile struct {
	ID               uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	UserId           uint   `json:"user_id"`
	ProfileFor       string `json:"profileFor"`
	Name             string `json:"name"`
	Image            string `json:"image"`
	DateOfBirth      string `json:"dateOfBirth"` // Updated to string to match the date format in the JSON
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
	Siblings         string `json:"siblings"` // Keeping this as string since it can be a number or a list
	Qualification    string `json:"qualification"`
	Occupation       string `json:"occupation"`
	AnnualIncome     string `json:"annualIncome"` // Updated to match JSON format
	MaritalStatus    string `json:"maritalStatus"`
	Address          string `json:"address"`
	CurrentLocation  string `json:"currentLocation"`
	Status           string `json:"status"`
	PhoneNumbers     string `json:"phoneNumbers"` // Keeping this as string for multiple numbers (could be CSV or JSON)
}
