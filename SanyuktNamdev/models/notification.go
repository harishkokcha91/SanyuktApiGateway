package models

import (
	"time"
)

type Notification struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `json:"user_id" gorm:"not null;index"`
	Type      string    `json:"type" gorm:"not null;size:50;index"`
	Message   string    `json:"message" gorm:"not null;type:text"`
	Channel   string    `json:"channel" gorm:"not null;size:20;default:'in-app'"`
	Status    string    `json:"status" gorm:"not null;size:20;default:'pending';index"`
	SentAt    *time.Time `json:"sent_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relation
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeApproval   NotificationType = "Approval"
	NotificationTypeRejection  NotificationType = "Rejection"
	NotificationTypeEdit       NotificationType = "Edit"
	NotificationTypeSystem     NotificationType = "System"
)

// NotificationChannel represents the delivery channel
type NotificationChannel string

const (
	ChannelEmail   NotificationChannel = "email"
	ChannelInApp   NotificationChannel = "in-app"
	ChannelPush    NotificationChannel = "push"
)

// NotificationStatus represents the delivery status
type NotificationStatus string

const (
	StatusPending NotificationStatus = "Pending"
	StatusSent    NotificationStatus = "Sent"
	StatusFailed  NotificationStatus = "Failed"
)

// TableName returns the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}