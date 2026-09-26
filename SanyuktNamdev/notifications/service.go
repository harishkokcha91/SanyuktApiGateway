package notifications

import (
	"context"
	"fmt"
	"log"
	"time"

	"SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"gorm.io/gorm"
)

// Service handles notification operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new notification service
func NewService() *Service {
	return &Service{
		db: database.DB,
	}
}

// NotificationConfig holds configuration for notification channels
type NotificationConfig struct {
	EmailFrom     string
	EmailUsername string
	EmailPassword string
	EmailHost     string
	EmailPort     int
	SendGridKey   string
}

// DefaultConfig returns default notification config from environment
func DefaultConfig() NotificationConfig {
	return NotificationConfig{
		EmailFrom:     "noreply@sanyukt.com",
		EmailUsername: "",
		EmailPassword: "",
		EmailHost:     "",
		EmailPort:     587,
		SendGridKey:   "",
	}
}

// CreateInAppNotification creates an in-app notification for a user
func (s *Service) CreateInAppNotification(userID uint, notifType, message string) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:  userID,
		Type:    notifType,
		Message: message,
		Channel: string(models.ChannelInApp),
		Status:  string(models.StatusPending),
	}

	if err := s.db.Create(notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create in-app notification: %w", err)
	}

	return notification, nil
}

// CreateEmailNotification creates an email notification and queues it for sending
func (s *Service) CreateEmailNotification(userID uint, notifType, message string) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:  userID,
		Type:    notifType,
		Message: message,
		Channel: string(models.ChannelEmail),
		Status:  string(models.StatusPending),
	}

	if err := s.db.Create(notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create email notification: %w", err)
	}

	return notification, nil
}

// QueueNotification queues a notification for a specific channel
func (s *Service) QueueNotification(userID uint, notifType string, channel models.NotificationChannel, message string) (*models.Notification, error) {
	switch channel {
	case models.ChannelInApp:
		return s.CreateInAppNotification(userID, notifType, message)
	case models.ChannelEmail:
		return s.CreateEmailNotification(userID, notifType, message)
	case models.ChannelPush:
		// Push notifications - placeholder for future implementation
		notification := &models.Notification{
			UserID:  userID,
			Type:    notifType,
			Message: message,
			Channel: string(models.ChannelPush),
			Status:  string(models.StatusPending),
		}
		if err := s.db.Create(notification).Error; err != nil {
			return nil, fmt.Errorf("failed to create push notification: %w", err)
		}
		return notification, nil
	default:
		return nil, fmt.Errorf("unsupported notification channel: %s", channel)
	}
}

// SendEmailNotification sends an email notification
// This is a placeholder - in production, integrate with SendGrid, SMTP, etc.
func (s *Service) SendEmailNotification(notification *models.Notification) error {
	// In a real implementation, this would:
	// 1. Fetch user email from database
	// 2. Send email via SMTP/SendGrid
	// 3. Update notification status

	var user models.User
	if err := s.db.First(&user, notification.UserID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Placeholder for actual email sending
	// For now, we'll just log and mark as sent
	log.Printf("EMAIL to %s: %s - %s", user.Email, notification.Type, notification.Message)

	// Update status to sent
	notification.Status = string(models.StatusSent)
	now := time.Now()
	notification.SentAt = &now
	if err := s.db.Save(notification).Error; err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

// ProcessPendingEmails processes all pending email notifications
func (s *Service) ProcessPendingEmails(ctx context.Context) error {
	var notifications []models.Notification
	if err := s.db.Where("channel = ? AND status = ?", models.ChannelEmail, models.StatusPending).Find(&notifications).Error; err != nil {
		return fmt.Errorf("failed to fetch pending emails: %w", err)
	}

	for _, notif := range notifications {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.SendEmailNotification(&notif); err != nil {
				log.Printf("Failed to send email notification %d: %v", notif.ID, err)
				notif.Status = string(models.StatusFailed)
				s.db.Save(&notif)
			}
		}
	}

	return nil
}

// GetInAppNotifications retrieves in-app notifications for a user
func (s *Service) GetInAppNotifications(userID uint, page, limit int, unreadOnly bool) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := s.db.Model(&models.Notification{}).Where("user_id = ? AND channel = ?", userID, models.ChannelInApp)

	if unreadOnly {
		query = query.Where("status != ?", models.StatusSent)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	return notifications, total, nil
}

// MarkAsRead marks a notification as read (sent)
func (s *Service) MarkAsRead(notificationID, userID uint) error {
	result := s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":  models.StatusSent,
			"sent_at": time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAllAsRead marks all in-app notifications as read for a user
func (s *Service) MarkAllAsRead(userID uint) error {
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND channel = ? AND status = ?", userID, models.ChannelInApp, models.StatusPending).
		Updates(map[string]interface{}{
			"status":  models.StatusSent,
			"sent_at": time.Now(),
		}).Error
}

// NotifyOnApproval sends approval notification to content owner
func (s *Service) NotifyOnApproval(userID uint, contentType string, contentName string) error {
	message := fmt.Sprintf("Your %s \"%s\" has been approved.", contentType, contentName)
	_, err := s.QueueNotification(userID, string(models.NotificationTypeApproval), models.ChannelInApp, message)
	if err != nil {
		return err
	}

	// Also queue email notification
	_, err = s.QueueNotification(userID, string(models.NotificationTypeApproval), models.ChannelEmail, message)
	return err
}

// NotifyOnRejection sends rejection notification to content owner
func (s *Service) NotifyOnRejection(userID uint, contentType string, contentName string, reason string) error {
	message := fmt.Sprintf("Your %s \"%s\" has been rejected.", contentType, contentName)
	if reason != "" {
		message += " Reason: " + reason
	}

	_, err := s.QueueNotification(userID, string(models.NotificationTypeRejection), models.ChannelInApp, message)
	if err != nil {
		return err
	}

	// Also queue email notification
	_, err = s.QueueNotification(userID, string(models.NotificationTypeRejection), models.ChannelEmail, message)
	return err
}

// NotifyOnOwnerEdit sends notification to admins when owner edits approved content
func (s *Service) NotifyOnOwnerEdit(contentType string, contentName string, ownerName string) error {
	message := fmt.Sprintf("A %s \"%s\" has been edited by %s and requires review.", contentType, contentName, ownerName)

	// Find all admin users
	var admins []models.User
	if err := s.db.Where("role = ?", "admin").Find(&admins).Error; err != nil {
		return fmt.Errorf("failed to find admin users: %w", err)
	}

	for _, admin := range admins {
		// In-app notification for admin
		if _, err := s.QueueNotification(admin.ID, string(models.NotificationTypeEdit), models.ChannelInApp, message); err != nil {
			log.Printf("Failed to create admin notification: %v", err)
		}

		// Email notification for admin
		if _, err := s.QueueNotification(admin.ID, string(models.NotificationTypeEdit), models.ChannelEmail, message); err != nil {
			log.Printf("Failed to create admin email notification: %v", err)
		}
	}

	return nil
}

// NotifySystem sends a system notification
func (s *Service) NotifySystem(userID uint, message string, channel models.NotificationChannel) error {
	_, err := s.QueueNotification(userID, string(models.NotificationTypeSystem), channel, message)
	return err
}