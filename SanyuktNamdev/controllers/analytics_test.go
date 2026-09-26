package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"SanyuktNamdev/middleware"
	"SanyuktNamdev/models"
	"SanyuktNamdev/testhelpers"
	"SanyuktNamdev/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupAnalyticsTestDB(t *testing.T) *gorm.DB {
	db := testhelpers.SetupTestDBWithModels(t,
		&models.User{},
		&models.Profile{},
		&models.Business{},
		&models.Event{},
		&models.Achievement{},
		&models.Notification{},
	)
	// Note: defer close handled by testhelpers

	// Create test users
	admin := testhelpers.CreateTestUser(t, db, 1, "admin@test.com", "Admin")
	admin.Role = "admin"
	db.Save(admin)

	user1 := testhelpers.CreateTestUser(t, db, 2, "user1@test.com", "User1")
	user2 := testhelpers.CreateTestUser(t, db, 3, "user2@test.com", "User2")

	// Create test profiles
	profiles := []models.Profile{
		{UserId: user1.ID, Name: "Profile 1", Status: "active", ProfileFor: "Self", DateOfBirth: "1990-01-01", PhoneNumbers: "1234567890", MaritalStatus: "Single", CurrentLocation: "NYC"},
		{UserId: user1.ID, Name: "Profile 2", Status: "pending", ProfileFor: "Self", DateOfBirth: "1991-01-01", PhoneNumbers: "1234567890", MaritalStatus: "Single", CurrentLocation: "LA"},
		{UserId: user2.ID, Name: "Profile 3", Status: "inactive", ProfileFor: "Self", DateOfBirth: "1992-01-01", PhoneNumbers: "1234567890", MaritalStatus: "Single", CurrentLocation: "Chicago"},
	}
	for i := range profiles {
		db.Create(&profiles[i])
		// Set approved_at for active profiles
		if profiles[i].Status == "active" {
			now := time.Now()
			db.Model(&profiles[i]).Update("approved_at", now)
		}
	}

	// Create test businesses
	businesses := []models.Business{
		{Name: "Business 1", Category: "Food", Status: "Approved", Owner: "Owner 1", Email: "b1@test.com", Location: "NYC", Address: "123 St", City: "NYC", State: "NY", ZipCode: "10001", Country: "USA"},
		{Name: "Business 2", Category: "Retail", Status: "Pending", Owner: "Owner 2", Email: "b2@test.com", Location: "LA", Address: "456 Ave", City: "LA", State: "CA", ZipCode: "90001", Country: "USA"},
		{Name: "Business 3", Category: "Service", Status: "Rejected", Owner: "Owner 3", Email: "b3@test.com", Location: "Chicago", Address: "789 Blvd", City: "Chicago", State: "IL", ZipCode: "60601", Country: "USA"},
	}
	for i := range businesses {
		db.Create(&businesses[i])
		if businesses[i].Status == "Approved" {
			now := time.Now()
			db.Model(&businesses[i]).Update("approved_at", now)
		}
	}

	// Create test events
	events := []models.Event{
		{Name: "Event 1", Status: "Approved", EventDate: "2024-12-01", Venue: "Venue 1", Organizer: "Org 1", Email: "e1@test.com", Category: "Conference", City: "NYC", State: "NY", ZipCode: "10001", Country: "USA"},
		{Name: "Event 2", Status: "Pending", EventDate: "2024-12-15", Venue: "Venue 2", Organizer: "Org 2", Email: "e2@test.com", Category: "Workshop", City: "LA", State: "CA", ZipCode: "90001", Country: "USA"},
		{Name: "Event 3", Status: "Rejected", EventDate: "2024-12-20", Venue: "Venue 3", Organizer: "Org 3", Email: "e3@test.com", Category: "Seminar", City: "Chicago", State: "IL", ZipCode: "60601", Country: "USA"},
	}
	for i := range events {
		db.Create(&events[i])
		if events[i].Status == "Approved" {
			now := time.Now()
			db.Model(&events[i]).Update("approved_at", now)
		}
	}

	// Create test achievements
	achievements := []models.Achievement{
		{Name: "Achievement 1", AchievementType: "Award", Status: "Approved", DateOfAchievement: "2024-01-01", Achievement: "Test achievement 1"},
		{Name: "Achievement 2", AchievementType: "Certification", Status: "Pending", DateOfAchievement: "2024-02-01", Achievement: "Test achievement 2"},
		{Name: "Achievement 3", AchievementType: "Recognition", Status: "Rejected", DateOfAchievement: "2024-03-01", Achievement: "Test achievement 3"},
	}
	for i := range achievements {
		db.Create(&achievements[i])
		if achievements[i].Status == "Approved" {
			now := time.Now()
			db.Model(&achievements[i]).Update("approved_at", now)
		}
	}

	return db
}

func setupTestRouter(db *gorm.DB, token string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userid", 1)
		c.Set("role", "admin")
		c.Next()
	})
	// Manually register admin analytics routes
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		adminGroup.GET("/analytics/summary", AdminAnalyticsSummary)
		adminGroup.GET("/analytics/pending", AdminAnalyticsPending)
		adminGroup.GET("/analytics/turnaround", AdminAnalyticsTurnaround)
		adminGroup.GET("/analytics/rejections", AdminAnalyticsRejections)
	}
	return r
}

func TestAnalytics_Summary(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")
	r := setupTestRouter(db, adminToken)

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/summary", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Analytics summary retrieved successfully", resp["message"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(12), data["total_items"])
	assert.Equal(t, float64(4), data["total_pending"])
	assert.Equal(t, float64(4), data["total_approved"])
	assert.Equal(t, float64(4), data["total_rejected"])
}

func TestAnalytics_Summary_NonAdminForbidden(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	userToken := testhelpers.GenerateTestToken(t, "2", "user")
	r := setupTestRouter(db, userToken)

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/summary", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAnalytics_Pending(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")
	r := setupTestRouter(db, adminToken)

	// Test profile pending
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/pending?type=profile", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	// Response doesn't include "message" field, check for data instead
	assert.Equal(t, float64(1), resp["totalRecords"])
	assert.Len(t, resp["data"], 1)
	assert.Equal(t, "profile", resp["data"].([]interface{})[0].(map[string]interface{})["type"])
	assert.Equal(t, "pending", resp["data"].([]interface{})[0].(map[string]interface{})["status"])

	// Test business pending
	req = httptest.NewRequest(http.MethodGet, "/admin/analytics/pending?type=business", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["totalRecords"])
	assert.Equal(t, "business", resp["data"].([]interface{})[0].(map[string]interface{})["type"])
	assert.Equal(t, "Pending", resp["data"].([]interface{})[0].(map[string]interface{})["status"])

	// Test pagination
	req = httptest.NewRequest(http.MethodGet, "/admin/analytics/pending?type=profile&page=1&limit=1", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["totalRecords"])
	assert.Len(t, resp["data"], 1)

	// Test missing type parameter
	req = httptest.NewRequest(http.MethodGet, "/admin/analytics/pending", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalytics_Pending_InvalidType(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")
	r := setupTestRouter(db, adminToken)

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/pending?type=invalid", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalytics_Turnaround(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")
	r := setupTestRouter(db, adminToken)

	// Test all types
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/turnaround", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Turnaround statistics retrieved successfully", resp["message"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 4)

	// Test specific type
	req = httptest.NewRequest(http.MethodGet, "/admin/analytics/turnaround?type=profile", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "profile", resp["data"].(map[string]interface{})["model_type"])
	assert.Equal(t, float64(1), resp["data"].(map[string]interface{})["total_approved"])
}

func TestAnalytics_Turnaround_InvalidType(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")
	r := setupTestRouter(db, adminToken)

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/turnaround?type=invalid", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnalytics_Rejections(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")
	r := setupTestRouter(db, adminToken)

	// Test all types
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/rejections", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Rejection rates retrieved successfully", resp["message"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 4)

	// Verify rejection rate calculation (1 rejected out of 3 = 33.33%)
	for _, item := range data {
		m := item.(map[string]interface{})
		assert.Equal(t, float64(3), m["total_created"])
		assert.Equal(t, float64(1), m["total_rejected"])
		assert.InDelta(t, 33.33, m["rejection_rate"], 0.01)
	}

	// Test specific type
	req = httptest.NewRequest(http.MethodGet, "/admin/analytics/rejections?type=business", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "business", resp["data"].(map[string]interface{})["model_type"])
}

func TestAnalytics_Rejections_InvalidType(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	adminToken := testhelpers.GenerateTestToken(t, "1", "admin")
	r := setupTestRouter(db, adminToken)

	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/rejections?type=invalid", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUtils_GetStatusCounts(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	counts, err := utils.GetStatusCounts("profile")
	require.NoError(t, err)
	assert.Len(t, counts, 3)

	found := make(map[string]int64)
	for _, c := range counts {
		found[c.Status] = c.Count
	}
	assert.Equal(t, int64(1), found["pending"])
	assert.Equal(t, int64(1), found["active"])
	assert.Equal(t, int64(1), found["inactive"])
}

func TestUtils_GetAllStatusCounts(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	summary, err := utils.GetAllStatusCounts()
	require.NoError(t, err)
	assert.Equal(t, int64(12), summary.TotalItems)
	assert.Equal(t, int64(4), summary.TotalPending)
	assert.Equal(t, int64(4), summary.TotalApproved)
	assert.Equal(t, int64(4), summary.TotalRejected)
}

func TestUtils_GetPendingItems(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	items, total, err := utils.GetPendingItems("profile", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
	assert.Equal(t, "pending", items[0].Status)
}

func TestUtils_GetApprovalTurnaround(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	stats, err := utils.GetApprovalTurnaround("profile")
	require.NoError(t, err)
	assert.Equal(t, "profile", stats.ModelType)
	assert.Equal(t, int64(1), stats.TotalApproved)
	assert.True(t, stats.AverageTurnaround >= 0)
}

func TestUtils_GetRejectionRates(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	rate, err := utils.GetRejectionRates("profile")
	require.NoError(t, err)
	assert.Equal(t, "profile", rate.ModelType)
	assert.Equal(t, int64(3), rate.TotalCreated)
	assert.Equal(t, int64(1), rate.TotalRejected)
	assert.InDelta(t, 33.33, rate.RejectionRate, 0.01)
}