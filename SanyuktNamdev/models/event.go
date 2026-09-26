package models

import "time"

type Event struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" binding:"required,min=3,max=200"`
	Description   string    `json:"description" binding:"max=2000"`
	EventDate     string    `json:"event_date" binding:"required,datetime=2006-01-02"`
	Venue         string    `json:"venue" binding:"required,max=200"`
	Address       string    `json:"address" binding:"max=300"`
	City          string    `json:"city" binding:"required,max=100"`
	State         string    `json:"state" binding:"required,max=100"`
	ZipCode       string    `json:"zip_code" binding:"max=20"`
	Country       string    `json:"country" binding:"required,max=100"`
	Organizer     string    `json:"organizer" binding:"required,max=100"`
	Email         string    `json:"email" binding:"required,email"`
	Phone         string    `json:"phone" binding:"max=20"`
	Category      string    `json:"category" binding:"required,max=50"`
	Capacity      string    `json:"capacity" binding:"max=20"`
	Attendees     string    `json:"attendees_registered" binding:"max=20"`
	Status        string    `json:"status" gorm:"default:'Pending'" binding:"omitempty,oneof=Pending Approved Upcoming Completed Cancelled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Image         string    `json:"image"`
	RegLink       string    `json:"registration_link" binding:"omitempty,url"`
	IsOnline      bool      `json:"is_online"`
	TicketPrice   string    `json:"ticket_price" binding:"max=20"`

	// Approval audit fields
	ApprovedBy  *uint      `json:"approved_by" gorm:"index"`
	ApprovedAt  *time.Time `json:"approved_at"`
	RejectedBy  *uint      `json:"rejected_by" gorm:"index"`
	RejectedAt  *time.Time `json:"rejected_at"`
}