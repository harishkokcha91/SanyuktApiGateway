package testhelpers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Profile{}, &models.Event{}, &models.Business{}, &models.Achievement{}, &models.Notification{}, &models.NotificationPreference{})
	require.NoError(t, err)

	database.DB = db
	return db
}

// SetupTestDBWithModels creates an in-memory SQLite database with specific models
func SetupTestDBWithModels(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(models...)
	require.NoError(t, err)

	database.DB = db
	return db
}

// GenerateTestToken generates a JWT token for testing
func GenerateTestToken(t *testing.T, userID string, role string) string {
	t.Helper()
	// Set test JWT secret if not set
	if len(utils.JWTSecretForTest()) == 0 {
		utils.SetTestJWTSecret("test-secret-key-that-is-at-least-32-bytes-long")
	}
	token, err := utils.GenerateToken(userID, role)
	require.NoError(t, err)
	return token
}

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, db *gorm.DB, id uint, email, name string) *models.User {
	t.Helper()
	user := &models.User{
		Password:     "hashedpassword",
		Email:        email,
		Name:         name,
		Age:          25,
		Status:       "active",
		Role:         "user",
		PhoneNumbers: "1234567890",
	}
	user.ID = id
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

// CreateTestProfile creates a test profile in the database
func CreateTestProfile(t *testing.T, db *gorm.DB, id uint, userID uint, name string) *models.Profile {
	t.Helper()
	profile := &models.Profile{
		UserId:          userID,
		ProfileFor:      "self",
		Name:            name,
		DateOfBirth:     "1990-01-01",
		MaritalStatus:   "Single",
		Status:          "pending",
		PhoneNumbers:    "1234567890",
		BirthPlace:      "Test City",
		Height:          "5'10",
		Complexion:      "Fair",
		FatherName:      "Father",
		MotherName:      "Mother",
		Siblings:        "1 brother",
		Qualification:   "B.Tech",
		Occupation:      "Engineer",
		AnnualIncome:    "10L",
		Address:         "123 Test St",
		CurrentLocation: "Test City",
	}
	profile.ID = id
	err := db.Create(profile).Error
	require.NoError(t, err)
	return profile
}

// CreateTestEvent creates a test event in the database
func CreateTestEvent(t *testing.T, db *gorm.DB, id uint) *models.Event {
	t.Helper()
	event := &models.Event{
		Name:        "Test Event",
		Description: "Test Description",
		EventDate:   "2025-12-31",
		Venue:       "Test Venue",
		City:        "Test City",
		State:       "Test State",
		Country:     "Test Country",
		Organizer:   "Test Organizer",
		Email:       "organizer@test.com",
		Category:    "Tech",
		Status:      "Approved",
	}
	event.ID = id
	err := db.Create(event).Error
	require.NoError(t, err)
	return event
}

// CreateTestBusiness creates a test business in the database
func CreateTestBusiness(t *testing.T, db *gorm.DB, id uint) *models.Business {
	t.Helper()
	business := &models.Business{
		Name:        "Test Business",
		Category:    "Restaurant",
		Description: "Test Description",
		Owner:       "Test Owner",
		Email:       "owner@test.com",
		Location:    "Test Location",
		City:        "Test City",
		State:       "Test State",
		Country:     "Test Country",
		Status:      "Approved",
	}
	business.ID = id
	err := db.Create(business).Error
	require.NoError(t, err)
	return business
}

// CreateTestAchievement creates a test achievement in the database
func CreateTestAchievement(t *testing.T, db *gorm.DB, id uint) *models.Achievement {
	t.Helper()
	achievement := &models.Achievement{
		Name:              "Test Achievement",
		AchievementType:   "Academic",
		Achievement:       "Test Achievement Detail",
		Description:       "Test Description",
		DateOfAchievement: "2025-01-01",
		Status:            "Approved",
	}
	achievement.ID = id
	err := db.Create(achievement).Error
	require.NoError(t, err)
	return achievement
}

// MakeJSONRequest makes a JSON HTTP request
func MakeJSONRequest(r *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// AssertStatus asserts the response status code
func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	require.Equal(t, expected, w.Code, "Response body: %s", w.Body.String())
}

// AssertJSONResponse asserts the response contains expected JSON fields
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedFields map[string]interface{}) {
	t.Helper()
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	for k, v := range expectedFields {
		require.Equal(t, v, resp[k], "Field %s mismatch", k)
	}
}

// AssertErrorResponse asserts the response is an error with expected message
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedCode int, containsMsg string) {
	t.Helper()
	AssertStatus(t, w, expectedCode)
	require.Contains(t, w.Body.String(), containsMsg)
}

// ParseTime parses a time string in RFC3339 format for test data
func ParseTime(t *testing.T, timeStr string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, timeStr)
	require.NoError(t, err)
	return parsed
}

// CreateTestNotification creates a test notification in the database
func CreateTestNotification(t *testing.T, db *gorm.DB, id uint, userID uint, notifType, message string) *models.Notification {
	t.Helper()
	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationType(notifType),
		Message: message,
		Channel: models.ChannelInApp,
		Status:  models.StatusPending,
	}
	notification.ID = id
	err := db.Create(notification).Error
	require.NoError(t, err)
	return notification
}