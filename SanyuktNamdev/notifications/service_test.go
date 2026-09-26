package notifications

import (
	"testing"
	"time"

	"SanyuktNamdev/models"
	"SanyuktNamdev/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateInAppNotification(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create test user
	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	notification, err := service.CreateInAppNotification(user.ID, string(models.NotificationTypeApproval), "Your profile has been approved.")
	require.NoError(t, err)
	assert.NotNil(t, notification)
	assert.Equal(t, user.ID, notification.UserID)
	assert.Equal(t, models.NotificationTypeApproval, notification.Type)
	assert.Equal(t, "Your profile has been approved.", notification.Message)
	assert.Equal(t, models.ChannelInApp, notification.Channel)
	assert.Equal(t, models.StatusPending, notification.Status)
	assert.Nil(t, notification.SentAt)

	// Verify it's in the database
	var saved models.Notification
	db.First(&saved, notification.ID)
	assert.Equal(t, notification.ID, saved.ID)
}

func TestCreateEmailNotification(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	notification, err := service.CreateEmailNotification(user.ID, string(models.NotificationTypeRejection), "Your profile was rejected.")
	require.NoError(t, err)
	assert.NotNil(t, notification)
	assert.Equal(t, user.ID, notification.UserID)
	assert.Equal(t, models.NotificationTypeRejection, notification.Type)
	assert.Equal(t, "Your profile was rejected.", notification.Message)
	assert.Equal(t, models.ChannelEmail, notification.Channel)
	assert.Equal(t, models.StatusPending, notification.Status)
	assert.Nil(t, notification.SentAt)
}

func TestQueueNotification_InApp(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	notification, err := service.QueueNotification(user.ID, string(models.NotificationTypeSystem), models.ChannelInApp, "Test message")
	require.NoError(t, err)
	assert.Equal(t, models.ChannelInApp, notification.Channel)
	assert.Equal(t, models.StatusPending, notification.Status)
}

func TestQueueNotification_Email(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	notification, err := service.QueueNotification(user.ID, string(models.NotificationTypeSystem), models.ChannelEmail, "Test message")
	require.NoError(t, err)
	assert.Equal(t, models.ChannelEmail, notification.Channel)
	assert.Equal(t, models.StatusPending, notification.Status)
}

func TestQueueNotification_Push(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	notification, err := service.QueueNotification(user.ID, string(models.NotificationTypeSystem), models.ChannelPush, "Test message")
	require.NoError(t, err)
	assert.Equal(t, models.ChannelPush, notification.Channel)
	assert.Equal(t, models.StatusPending, notification.Status)
}

func TestQueueNotification_InvalidChannel(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	_, err := service.QueueNotification(user.ID, string(models.NotificationTypeSystem), "invalid", "Test message")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported notification channel")
}

func TestGetInAppNotifications(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	user2 := testhelpers.CreateTestUser(t, db, 2, "other@test.com", "Other")

	service := NewService()

	// Create some notifications
	service.CreateInAppNotification(user.ID, string(models.NotificationTypeApproval), "Notification 1")
	time.Sleep(10 * time.Millisecond)
	service.CreateInAppNotification(user.ID, string(models.NotificationTypeRejection), "Notification 2")
	time.Sleep(10 * time.Millisecond)
	service.CreateInAppNotification(user.ID, string(models.NotificationTypeApproval), "Notification 3")
	service.CreateInAppNotification(user2.ID, string(models.NotificationTypeSystem), "Other user notification")

	// Get notifications for user 1
	notifications, total, err := service.GetInAppNotifications(user.ID, 1, 10, false)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, notifications, 3)

	// Should be ordered by created_at DESC (newest first)
	assert.Equal(t, "Notification 3", notifications[0].Message)
	assert.Equal(t, "Notification 2", notifications[1].Message)
	assert.Equal(t, "Notification 1", notifications[2].Message)
}

func TestGetInAppNotifications_UnreadOnly(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	// Create notifications with different statuses
	n1, _ := service.CreateInAppNotification(user.ID, string(models.NotificationTypeApproval), "Read notification")
	n1.Status = models.StatusSent
	db.Save(n1)

	service.CreateInAppNotification(user.ID, string(models.NotificationTypeRejection), "Unread notification 1")
	service.CreateInAppNotification(user.ID, string(models.NotificationTypeApproval), "Unread notification 2")

	// Get unread only
	notifications, total, err := service.GetInAppNotifications(user.ID, 1, 10, true)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, notifications, 2)
	for _, n := range notifications {
		assert.Equal(t, models.StatusPending, n.Status)
	}
}

func TestGetInAppNotifications_Pagination(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	// Create 5 notifications
	for i := 1; i <= 5; i++ {
		service.CreateInAppNotification(user.ID, string(models.NotificationTypeSystem), "Notification "+string(rune(i+'0')))
	}

	// Page 1, limit 2
	notifications, total, err := service.GetInAppNotifications(user.ID, 1, 2, false)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, notifications, 2)

	// Page 2, limit 2
	notifications, _, err = service.GetInAppNotifications(user.ID, 2, 2, false)
	require.NoError(t, err)
	assert.Len(t, notifications, 2)

	// Page 3, limit 2
	notifications, _, err = service.GetInAppNotifications(user.ID, 3, 2, false)
	require.NoError(t, err)
	assert.Len(t, notifications, 1)
}

func TestMarkAsRead(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	notification, _ := service.CreateInAppNotification(user.ID, string(models.NotificationTypeApproval), "Test notification")
	assert.Equal(t, models.StatusPending, notification.Status)

	// Mark as read
	err := service.MarkAsRead(notification.ID, user.ID)
	require.NoError(t, err)

	var saved models.Notification
	db.First(&saved, notification.ID)
	assert.Equal(t, models.StatusSent, saved.Status)
	assert.NotNil(t, saved.SentAt)
}

func TestMarkAsRead_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	err := service.MarkAsRead(999, user.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestMarkAllAsRead(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	// Create 3 notifications
	service.CreateInAppNotification(user.ID, string(models.NotificationTypeApproval), "Notification 1")
	service.CreateInAppNotification(user.ID, string(models.NotificationTypeRejection), "Notification 2")
	service.CreateInAppNotification(user.ID, string(models.NotificationTypeSystem), "Notification 3")

	// Mark all as read
	err := service.MarkAllAsRead(user.ID)
	require.NoError(t, err)

	// Verify all are marked as sent
	var count int64
	db.Model(&models.Notification{}).Where("user_id = ? AND channel = ? AND status = ?", user.ID, models.ChannelInApp, models.StatusPending).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestNotifyOnApproval(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	err := service.NotifyOnApproval(user.ID, "Profile", "My Profile")
	require.NoError(t, err)

	// Should create both in-app and email notifications
	var count int64
	db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", user.ID, models.NotificationTypeApproval).Count(&count)
	assert.Equal(t, int64(2), count) // 1 in-app + 1 email
}

func TestNotifyOnRejection(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	err := service.NotifyOnRejection(user.ID, "Business", "My Business", "Invalid address")
	require.NoError(t, err)

	var count int64
	db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", user.ID, models.NotificationTypeRejection).Count(&count)
	assert.Equal(t, int64(2), count) // 1 in-app + 1 email

	// Check message contains reason
	var notif models.Notification
	db.Where("user_id = ? AND type = ? AND channel = ?", user.ID, models.NotificationTypeRejection, models.ChannelInApp).First(&notif)
	assert.Contains(t, notif.Message, "Invalid address")
}

func TestNotifyOnOwnerEdit(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create admin user
	admin := testhelpers.CreateTestUser(t, db, 1, "admin@test.com", "Admin")
	admin.Role = "admin"
	db.Save(admin)

	// Create regular user (needed for context but not directly used)
	testhelpers.CreateTestUser(t, db, 2, "user@test.com", "User")

	service := NewService()

	err := service.NotifyOnOwnerEdit("Event", "Tech Conference", "User")
	require.NoError(t, err)

	// Should create notifications for admin (both in-app and email)
	var count int64
	db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", admin.ID, models.NotificationTypeEdit).Count(&count)
	assert.Equal(t, int64(2), count)

	// Check message content
	var notif models.Notification
	db.Where("user_id = ? AND type = ? AND channel = ?", admin.ID, models.NotificationTypeEdit, models.ChannelInApp).First(&notif)
	assert.Contains(t, notif.Message, "Tech Conference")
	assert.Contains(t, notif.Message, "User")
	assert.Contains(t, notif.Message, "edited")
}

func TestNotifyOnOwnerEdit_NoAdmins(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// No admin users
	testhelpers.CreateTestUser(t, db, 1, "user@test.com", "User")

	service := NewService()

	// Should not error even with no admins
	err := service.NotifyOnOwnerEdit("Profile", "My Profile", "User")
	require.NoError(t, err)
}

func TestSendEmailNotification(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{}, &models.NotificationPreference{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	service := NewService()

	// Create pending email notification
	notification, _ := service.CreateEmailNotification(user.ID, string(models.NotificationTypeApproval), "Test message")

	// Send it
	err := service.SendEmailNotification(notification)
	require.NoError(t, err)

	// Verify status updated
	var saved models.Notification
	db.First(&saved, notification.ID)
	assert.Equal(t, models.StatusSent, saved.Status)
	assert.NotNil(t, saved.SentAt)
}

func TestNotificationModel_TableName(t *testing.T) {
	var n models.Notification
	assert.Equal(t, "notifications", n.TableName())
}

func TestNotificationConstants(t *testing.T) {
	assert.Equal(t, "approval", string(models.NotificationTypeApproval))
	assert.Equal(t, "rejection", string(models.NotificationTypeRejection))
	assert.Equal(t, "edit", string(models.NotificationTypeEdit))
	assert.Equal(t, "system", string(models.NotificationTypeSystem))

	assert.Equal(t, "email", string(models.ChannelEmail))
	assert.Equal(t, "in-app", string(models.ChannelInApp))
	assert.Equal(t, "push", string(models.ChannelPush))

	assert.Equal(t, "pending", string(models.StatusPending))
	assert.Equal(t, "sent", string(models.StatusSent))
	assert.Equal(t, "failed", string(models.StatusFailed))
}