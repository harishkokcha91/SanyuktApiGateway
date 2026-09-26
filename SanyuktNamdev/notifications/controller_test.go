package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"SanyuktNamdev/models"
	"SanyuktNamdev/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestRouter(db *gorm.DB, service *Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r, service)
	return r
}

func TestController_GetNotifications(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	service := NewService()
	// Create some notifications
	service.CreateInAppNotification(user.ID, "Approval", "Notification 1")
	service.CreateInAppNotification(user.ID, "Rejection", "Notification 2")

	r := setupTestRouter(db, service)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(2), resp["totalRecords"])
	assert.Len(t, resp["data"], 2)
}

func TestController_GetNotifications_UnreadOnly(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	service := NewService()
	// Create read notification
	n1, _ := service.CreateInAppNotification(user.ID, "Approval", "Read notification")
	n1.Status = string(models.StatusSent)
	db.Save(n1)

	// Create unread notifications
	service.CreateInAppNotification(user.ID, "Rejection", "Unread 1")
	service.CreateInAppNotification(user.ID, "Approval", "Unread 2")

	r := setupTestRouter(db, service)

	req := httptest.NewRequest(http.MethodGet, "/notifications?unread_only=true", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(2), resp["totalRecords"])
}

func TestController_GetNotifications_NoAuth(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	service := NewService()
	r := setupTestRouter(db, service)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestController_GetUnreadCount(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	service := NewService()
	service.CreateInAppNotification(user.ID, "Approval", "Notification 1")
	service.CreateInAppNotification(user.ID, "Rejection", "Notification 2")

	r := setupTestRouter(db, service)

	req := httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(2), resp["unread_count"])
}

func TestController_MarkAsRead(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	service := NewService()
	notification, _ := service.CreateInAppNotification(user.ID, "Approval", "Test notification")

	r := setupTestRouter(db, service)

	// Test /notifications/:id/read
	req := httptest.NewRequest(http.MethodPatch, "/notifications/"+string(rune(notification.ID+'0'))+"/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Notification
	db.First(&saved, notification.ID)
	assert.Equal(t, string(models.StatusSent), saved.Status)
}

func TestController_MarkAsRead_MarkReadAlias(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	service := NewService()
	notification, _ := service.CreateInAppNotification(user.ID, "Approval", "Test notification")

	r := setupTestRouter(db, service)

	// Test /notifications/:id/mark-read (alias)
	req := httptest.NewRequest(http.MethodPatch, "/notifications/"+string(rune(notification.ID+'0'))+"/mark-read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Notification
	db.First(&saved, notification.ID)
	assert.Equal(t, string(models.StatusSent), saved.Status)
}

func TestController_MarkAsRead_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	token := testhelpers.GenerateTestToken(t, "1", "user")

	service := NewService()
	r := setupTestRouter(db, service)

	req := httptest.NewRequest(http.MethodPatch, "/notifications/999/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestController_MarkAllAsRead(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	service := NewService()
	service.CreateInAppNotification(user.ID, "Approval", "Notification 1")
	service.CreateInAppNotification(user.ID, "Rejection", "Notification 2")
	service.CreateInAppNotification(user.ID, "System", "Notification 3")

	r := setupTestRouter(db, service)

	req := httptest.NewRequest(http.MethodPatch, "/notifications/read-all", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	db.Model(&models.Notification{}).Where("user_id = ? AND channel = ? AND status = ?", user.ID, models.ChannelInApp, models.StatusPending).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestController_SendTestNotification_AdminSuccess(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create admin user
	admin := testhelpers.CreateTestUser(t, db, 1, "admin@test.com", "Admin")
	admin.Role = "admin"
	db.Save(admin)

	// Create target user
	targetUser := testhelpers.CreateTestUser(t, db, 2, "user@test.com", "User")
	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")

	service := NewService()
	r := setupTestRouter(db, service)

	body := map[string]interface{}{
		"user_id":  targetUser.ID,
		"type":     "System",
		"message":  "Test notification",
		"channel":  "in-app",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/notifications/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Test notification sent", resp["message"])
	assert.NotNil(t, resp["notification"])

	// Verify notification was created in DB
	var count int64
	db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", targetUser.ID, "System").Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestController_SendTestNotification_NonAdminForbidden(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	token := testhelpers.GenerateTestToken(t, "1", "user")

	targetUser := testhelpers.CreateTestUser(t, db, 2, "target@test.com", "Target")

	service := NewService()
	r := setupTestRouter(db, service)

	body := map[string]interface{}{
		"user_id":  targetUser.ID,
		"type":     "System",
		"message":  "Test notification",
		"channel":  "in-app",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/notifications/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestController_SendTestNotification_InvalidChannel(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	admin := testhelpers.CreateTestUser(t, db, 1, "admin@test.com", "Admin")
	admin.Role = "admin"
	db.Save(admin)

	targetUser := testhelpers.CreateTestUser(t, db, 2, "user@test.com", "User")
	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")

	service := NewService()
	r := setupTestRouter(db, service)

	body := map[string]interface{}{
		"user_id":  targetUser.ID,
		"type":     "System",
		"message":  "Test notification",
		"channel":  "invalid",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/notifications/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestController_SendTestNotification_MissingFields(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Profile{}, &models.Notification{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	admin := testhelpers.CreateTestUser(t, db, 1, "admin@test.com", "Admin")
	admin.Role = "admin"
	db.Save(admin)

	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")

	service := NewService()
	r := setupTestRouter(db, service)

	body := map[string]interface{}{
		"type":     "System",
		"message":  "Test notification",
		"channel":  "in-app",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/notifications/test", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}