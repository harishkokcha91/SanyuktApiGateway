package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"SanyuktNamdev/middleware"
	"SanyuktNamdev/models"
	"SanyuktNamdev/notifications"
	"SanyuktNamdev/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupNotificationPreferenceTestDB(t *testing.T) *gorm.DB {
	db := testhelpers.SetupTestDBWithModels(t,
		&models.User{},
		&models.NotificationPreference{},
		&models.Notification{},
	)
	// Note: defer close handled by testhelpers

	// Create test user
	testhelpers.CreateTestUser(t, db, 1, "user@test.com", "Test User")

	return db
}

func setupTestRouterWithAuth(db *gorm.DB, userID uint, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userid", userID)
		c.Set("role", role)
		c.Next()
	})
	prefController := NewNotificationPreferenceController()
	prefGroup := r.Group("/notification-preferences")
	prefGroup.Use(middleware.AuthMiddleware())
	{
		prefGroup.GET("", prefController.GetNotificationPreferences)
		prefGroup.PATCH("/:channel", prefController.UpdateNotificationPreference)
		prefGroup.POST("/reset", prefController.ResetNotificationPreferences)
	}
	return r
}

func TestNotificationPreferences_Get(t *testing.T) {
	db := setupNotificationPreferenceTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupTestRouterWithAuth(db, 1, "user")

	req := httptest.NewRequest(http.MethodGet, "/notification-preferences", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Notification preferences retrieved successfully", resp["message"])
	assert.Len(t, resp["data"], 3) // in-app, email, push

	// Verify all three channels exist
	channels := make(map[string]bool)
	for _, pref := range resp["data"].([]interface{}) {
		m := pref.(map[string]interface{})
		channels[m["channel"].(string)] = m["enabled"].(bool)
	}
	assert.True(t, channels["in-app"])
	assert.True(t, channels["email"])
	assert.True(t, channels["push"])
}

func TestNotificationPreferences_Get_NoAuth(t *testing.T) {
	db := setupNotificationPreferenceTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	r := setupTestRouterWithAuth(db, 1, "user")

	req := httptest.NewRequest(http.MethodGet, "/notification-preferences", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNotificationPreferences_Update(t *testing.T) {
	db := setupNotificationPreferenceTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupTestRouterWithAuth(db, 1, "user")

	// Disable email notifications
	body := map[string]interface{}{
		"enabled": false,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/notification-preferences/email", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Notification preference updated successfully", resp["message"])
	assert.Equal(t, "email", resp["data"].(map[string]interface{})["channel"])
	assert.Equal(t, false, resp["data"].(map[string]interface{})["enabled"])
}

func TestNotificationPreferences_Update_InvalidChannel(t *testing.T) {
	db := setupNotificationPreferenceTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupTestRouterWithAuth(db, 1, "user")

	body := map[string]interface{}{
		"enabled": false,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/notification-preferences/invalid", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationPreferences_Update_MissingEnabled(t *testing.T) {
	db := setupNotificationPreferenceTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupTestRouterWithAuth(db, 1, "user")

	body := map[string]interface{}{}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPatch, "/notification-preferences/email", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationPreferences_Reset(t *testing.T) {
	db := setupNotificationPreferenceTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupTestRouterWithAuth(db, 1, "user")

	// First, disable email and push
	prefController := NewNotificationPreferenceController()
	prefController.db.Exec("UPDATE notification_preferences SET enabled = false WHERE user_id = 1 AND channel IN ('email', 'push')")

	// Reset to defaults
	req := httptest.NewRequest(http.MethodPost, "/notification-preferences/reset", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Notification preferences reset to defaults successfully", resp["message"])
	assert.Len(t, resp["data"], 3)

	// Verify all are enabled
	for _, pref := range resp["data"].([]interface{}) {
		m := pref.(map[string]interface{})
		assert.True(t, m["enabled"].(bool), "Channel %s should be enabled after reset", m["channel"])
	}
}

func TestNotificationPreferences_IntegrationWithNotificationService(t *testing.T) {
	db := setupNotificationPreferenceTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create default preferences for user (enabled=true by default)
	for _, channel := range []string{"in-app", "email", "push"} {
		db.Exec("INSERT INTO notification_preferences (user_id, channel, enabled, created_at, updated_at) VALUES (?, ?, true, datetime('now'), datetime('now'))", 1, channel)
	}

	// Disable email for user
	db.Exec("UPDATE notification_preferences SET enabled = false WHERE user_id = 1 AND channel = 'email'")

	// Create notification service and try to queue email
	notifService := notifications.NewService()

	// Queue in-app notification (should work)
	inAppNotif, err := notifService.QueueNotification(1, "test", models.ChannelInApp, "Test message")
	require.NoError(t, err)
	assert.NotNil(t, inAppNotif)
	assert.Equal(t, models.ChannelInApp, inAppNotif.Channel)

	// Queue email notification (should be skipped due to preference)
	emailNotif, err := notifService.QueueNotification(1, "test", models.ChannelEmail, "Test message")
	require.NoError(t, err)
	assert.Nil(t, emailNotif, "Email notification should be nil when disabled")

	// Queue push notification (should work)
	pushNotif, err := notifService.QueueNotification(1, "test", models.ChannelPush, "Test message")
	require.NoError(t, err)
	assert.NotNil(t, pushNotif)
	assert.Equal(t, models.ChannelPush, pushNotif.Channel)
}

func TestNotificationPreferences_DefaultEnabledWhenNotSet(t *testing.T) {
	db := setupNotificationPreferenceTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Delete all preferences for user
	db.Exec("DELETE FROM notification_preferences WHERE user_id = 1")

	notifService := notifications.NewService()

	// All channels should be enabled by default
	channels := []models.NotificationChannel{
		models.ChannelInApp,
		models.ChannelEmail,
		models.ChannelPush,
	}

	for _, channel := range channels {
		notif, err := notifService.QueueNotification(1, "test", channel, "Test message")
		require.NoError(t, err)
		assert.NotNil(t, notif, "Channel %s should be enabled by default", channel)
	}
}