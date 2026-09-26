package models

import "time"

type Notification struct {
	ID        uint              `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint              `json:"user_id" gorm:"not null;index"`
	Type      NotificationType  `json:"type" gorm:"not null;size:50;index"`
	Message   string            `json:"message" gorm:"not null;type:text"`
	Channel   NotificationChannel `json:"channel" gorm:"not null;size:20;default:'in-app'"`
	Status    NotificationStatus  `json:"status" gorm:"not null;size:20;default:'pending';index"`
	SentAt    *time.Time        `json:"sent_at"`
	CreatedAt time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime" json:"updated_at"`

	// Relation
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeApproval  NotificationType = "approval"
	NotificationTypeRejection NotificationType = "rejection"
	NotificationTypeEdit      NotificationType = "edit"
	NotificationTypeSystem    NotificationType = "system"
)

// NotificationChannel represents the delivery channel
type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelInApp NotificationChannel = "in-app"
	ChannelPush  NotificationChannel = "push"
)

// NotificationStatus represents the delivery status
type NotificationStatus string

const (
	StatusPending NotificationStatus = "pending"
	StatusSent    NotificationStatus = "sent"
	StatusFailed  NotificationStatus = "failed"
)

// TableName returns the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	ID        uint              `gorm:"primaryKey" json:"id"`
	UserID    uint              `gorm:"not null;index" json:"user_id"`
	Channel   NotificationChannel `gorm:"not null;type:varchar(20)" json:"channel"`
	Enabled   bool              `gorm:"not null;default:true" json:"enabled"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// TableName specifies the table name for NotificationPreference
func (NotificationPreference) TableName() string {
	return "notification_preferences"
}