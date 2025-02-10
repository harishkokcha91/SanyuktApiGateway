package models

import "time"

type Event struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	EventDate   string    `json:"event_date"`
	Venue       string    `json:"venue"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	ZipCode     string    `json:"zip_code"`
	Country     string    `json:"country"`
	Organizer   string    `json:"organizer"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Category    string    `json:"category"`
	Capacity    string    `json:"capacity"`
	Attendees   string    `json:"attendees_registered"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Image       string    `json:"image"`
	RegLink     string    `json:"registration_link"`
	IsOnline    bool      `json:"is_online"`
	TicketPrice string    `json:"ticket_price"`
}
