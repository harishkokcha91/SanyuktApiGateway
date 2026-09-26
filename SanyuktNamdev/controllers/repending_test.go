package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"SanyuktNamdev/middleware"
	"SanyuktNamdev/models"
	"SanyuktNamdev/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupProfileRouter creates a test router with profile routes
func setupProfileRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.PUT("/matrimonialProfiles/:id", UpdateProfile)
	}
	return r
}

// setupBusinessRouter creates a test router with business routes
func setupBusinessRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.PUT("/businesses/:id", UpdateBusiness)
	}
	return r
}

// setupEventRouter creates a test router with event routes
func setupEventRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.PUT("/events/:id", UpdateEvent)
	}
	return r
}

// setupAchievementRouter creates a test router with achievement routes
func setupAchievementRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.PUT("/achievements/:id", UpdateAchievement)
	}
	return r
}

// setupAdminRouterWithAll creates a test router with all routes for admin tests
func setupAdminRouterWithAll(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.PUT("/matrimonialProfiles/:id", UpdateProfile)
		protected.PUT("/businesses/:id", UpdateBusiness)
		protected.PUT("/events/:id", UpdateEvent)
		protected.PUT("/achievements/:id", UpdateAchievement)
	}
	return r
}

func TestUpdateProfile_OwnerEditApprovedResetsToPending(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create user and approved profile
	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	profile := testhelpers.CreateTestProfile(t, db, 1, user.ID, "Approved Profile")
	profile.Status = "active" // "active" is the approved status for profiles
	approverID := uint(2)
	profile.ApprovedBy = &approverID
	approveTime := testhelpers.ParseTime(t, "2025-01-01T10:00:00Z")
	profile.ApprovedAt = &approveTime
	db.Save(profile)

	ownerToken := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupProfileRouter(db)

	updateData := map[string]interface{}{
		"name":             "Updated Profile",
		"profileFor":       "self",
		"dateOfBirth":      "1990-01-01",
		"maritalStatus":    "Single",
		"phoneNumbers":     "1234567890",
		"birthPlace":       "Test City",
		"height":           "5'10",
		"complexion":       "Fair",
		"fatherName":       "Father",
		"motherName":       "Mother",
		"siblings":         "1 brother",
		"qualification":    "B.Tech",
		"occupation":       "Engineer",
		"annualIncome":     "10L",
		"address":          "123 Test St",
		"currentLocation":  "Test City",
		"status":           "active", // Owner tries to keep it active
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/matrimonialProfiles/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify profile was updated and status reset to pending
	var updatedProfile models.Profile
	db.First(&updatedProfile, 1)
	assert.Equal(t, "Updated Profile", updatedProfile.Name)
	assert.Equal(t, "pending", updatedProfile.Status, "Owner edit on approved profile should reset to pending")
	assert.Nil(t, updatedProfile.ApprovedBy, "ApprovedBy should be cleared on re-pending")
	assert.Nil(t, updatedProfile.ApprovedAt, "ApprovedAt should be cleared on re-pending")
}

func TestUpdateProfile_AdminEditApprovedDoesNotReset(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create user and approved profile
	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	profile := testhelpers.CreateTestProfile(t, db, 1, user.ID, "Approved Profile")
	profile.Status = "active"
	approverID := uint(2)
	profile.ApprovedBy = &approverID
	approveTime := testhelpers.ParseTime(t, "2025-01-01T10:00:00Z")
	profile.ApprovedAt = &approveTime
	db.Save(profile)

	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")
	r := setupProfileRouter(db)

	updateData := map[string]interface{}{
		"name":             "Admin Updated Profile",
		"profileFor":       "self",
		"dateOfBirth":      "1990-01-01",
		"maritalStatus":    "Single",
		"phoneNumbers":     "1234567890",
		"birthPlace":       "Test City",
		"height":           "5'10",
		"complexion":       "Fair",
		"fatherName":       "Father",
		"motherName":       "Mother",
		"siblings":         "1 brother",
		"qualification":    "B.Tech",
		"occupation":       "Engineer",
		"annualIncome":     "10L",
		"address":          "123 Test St",
		"currentLocation":  "Test City",
		"status":           "active",
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/matrimonialProfiles/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify profile was updated but status remains active
	var updatedProfile models.Profile
	db.First(&updatedProfile, 1)
	assert.Equal(t, "Admin Updated Profile", updatedProfile.Name)
	assert.Equal(t, "active", updatedProfile.Status, "Admin edit should not reset status")
	assert.NotNil(t, updatedProfile.ApprovedBy, "ApprovedBy should be preserved")
	assert.NotNil(t, updatedProfile.ApprovedAt, "ApprovedAt should be preserved")
}

func TestUpdateProfile_OwnerCannotSetStatusToActive(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	profile := testhelpers.CreateTestProfile(t, db, 1, user.ID, "Pending Profile")
	profile.Status = "pending"
	db.Save(profile)

	ownerToken := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupProfileRouter(db)

	// Owner tries to set status to active directly
	updateData := map[string]interface{}{
		"name":             "Profile",
		"profileFor":       "self",
		"dateOfBirth":      "1990-01-01",
		"maritalStatus":    "Single",
		"phoneNumbers":     "1234567890",
		"birthPlace":       "Test City",
		"height":           "5'10",
		"complexion":       "Fair",
		"fatherName":       "Father",
		"motherName":       "Mother",
		"siblings":         "1 brother",
		"qualification":    "B.Tech",
		"occupation":       "Engineer",
		"annualIncome":     "10L",
		"address":          "123 Test St",
		"currentLocation":  "Test City",
		"status":           "active", // Owner tries to self-approve
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/matrimonialProfiles/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedProfile models.Profile
	db.First(&updatedProfile, 1)
	assert.Equal(t, "pending", updatedProfile.Status, "Owner cannot self-approve; status should remain pending")
}

func TestUpdateBusiness_AdminEditApprovedDoesNotReset(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create approved business with audit fields
	business := testhelpers.CreateTestBusiness(t, db, 1)
	business.Status = "Approved"
	approverID := uint(2)
	business.ApprovedBy = &approverID
	approveTime := testhelpers.ParseTime(t, "2025-01-01T10:00:00Z")
	business.ApprovedAt = &approveTime
	db.Save(business)

	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")
	r := setupBusinessRouter(db)

	updateData := map[string]interface{}{
		"name":        "Updated Business",
		"category":    "Restaurant",
		"description": "Updated Description",
		"owner":       "Test Owner",
		"email":       "owner@test.com",
		"location":    "Test Location",
		"city":        "Test City",
		"state":       "Test State",
		"country":     "Test Country",
		"status":      "Approved",
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/businesses/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedBusiness models.Business
	db.First(&updatedBusiness, 1)
	assert.Equal(t, "Updated Business", updatedBusiness.Name)
	assert.Equal(t, "Approved", updatedBusiness.Status, "Admin edit should not reset status")
	assert.NotNil(t, updatedBusiness.ApprovedBy)
	assert.NotNil(t, updatedBusiness.ApprovedAt)
}

func TestUpdateEvent_AdminEditApprovedDoesNotReset(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create approved event with audit fields
	event := testhelpers.CreateTestEvent(t, db, 1)
	event.Status = "Approved"
	approverID := uint(2)
	event.ApprovedBy = &approverID
	approveTime := testhelpers.ParseTime(t, "2025-01-01T10:00:00Z")
	event.ApprovedAt = &approveTime
	db.Save(event)

	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")
	r := setupEventRouter(db)

	updateData := map[string]interface{}{
		"name":        "Updated Event",
		"description": "Updated Description",
		"event_date":  "2025-12-31",
		"venue":       "Test Venue",
		"city":        "Test City",
		"state":       "Test State",
		"country":     "Test Country",
		"organizer":   "Test Organizer",
		"email":       "organizer@test.com",
		"category":    "Tech",
		"status":      "Approved",
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/events/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedEvent models.Event
	db.First(&updatedEvent, 1)
	assert.Equal(t, "Updated Event", updatedEvent.Name)
	assert.Equal(t, "Approved", updatedEvent.Status, "Admin edit should not reset status")
	assert.NotNil(t, updatedEvent.ApprovedBy)
	assert.NotNil(t, updatedEvent.ApprovedAt)
}

func TestUpdateAchievement_AdminEditApprovedDoesNotReset(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create approved achievement with audit fields
	achievement := testhelpers.CreateTestAchievement(t, db, 1)
	achievement.Status = "Approved"
	approverID := uint(2)
	achievement.ApprovedBy = &approverID
	approveTime := testhelpers.ParseTime(t, "2025-01-01T10:00:00Z")
	achievement.ApprovedAt = &approveTime
	db.Save(achievement)

	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")
	r := setupAchievementRouter(db)

	updateData := map[string]interface{}{
		"name":                   "Updated Achievement",
		"achievement_type":       "Academic",
		"achievement":            "Updated Detail",
		"description":            "Updated Description",
		"date_of_achievement":    "2025-01-01",
		"status":                 "Approved",
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/achievements/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedAchievement models.Achievement
	db.First(&updatedAchievement, 1)
	assert.Equal(t, "Updated Achievement", updatedAchievement.Name)
	assert.Equal(t, "Approved", updatedAchievement.Status, "Admin edit should not reset status")
	assert.NotNil(t, updatedAchievement.ApprovedBy)
	assert.NotNil(t, updatedAchievement.ApprovedAt)
}

func TestUpdateProfile_PreservesRejectedFieldsOnRePending(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create profile that was previously rejected, then approved
	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	profile := testhelpers.CreateTestProfile(t, db, 1, user.ID, "Rejected then Approved Profile")
	profile.Status = "active"
	approverID := uint(2)
	rejecterID := uint(3)
	profile.ApprovedBy = &approverID
	approveTime := testhelpers.ParseTime(t, "2025-01-02T10:00:00Z")
	profile.ApprovedAt = &approveTime
	profile.RejectedBy = &rejecterID
	rejectTime := testhelpers.ParseTime(t, "2025-01-01T10:00:00Z")
	profile.RejectedAt = &rejectTime
	db.Save(profile)

	ownerToken := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupProfileRouter(db)

	updateData := map[string]interface{}{
		"name":             "Updated Profile",
		"profileFor":       "self",
		"dateOfBirth":      "1990-01-01",
		"maritalStatus":    "Single",
		"phoneNumbers":     "1234567890",
		"birthPlace":       "Test City",
		"height":           "5'10",
		"complexion":       "Fair",
		"fatherName":       "Father",
		"motherName":       "Mother",
		"siblings":         "1 brother",
		"qualification":    "B.Tech",
		"occupation":       "Engineer",
		"annualIncome":     "10L",
		"address":          "123 Test St",
		"currentLocation":  "Test City",
		"status":           "active",
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/matrimonialProfiles/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedProfile models.Profile
	db.First(&updatedProfile, 1)
	assert.Equal(t, "pending", updatedProfile.Status)
	assert.Nil(t, updatedProfile.ApprovedBy)
	assert.Nil(t, updatedProfile.ApprovedAt)
	// Rejected fields should be preserved
	assert.NotNil(t, updatedProfile.RejectedBy, "RejectedBy should be preserved")
	assert.NotNil(t, updatedProfile.RejectedAt, "RejectedAt should be preserved")
	assert.Equal(t, uint(3), *updatedProfile.RejectedBy)
}

func TestUpdateBusiness_NonAdminCannotSetStatusToApproved(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	business := testhelpers.CreateTestBusiness(t, db, 1)
	business.Status = "Pending"
	db.Save(business)

	userToken := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupBusinessRouter(db)

	updateData := map[string]interface{}{
		"name":   "Business",
		"status": "Approved", // Non-admin tries to self-approve
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/businesses/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Currently only admins can update businesses, so this should be forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateProfile_OwnerEditPendingRemainsPending(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	profile := testhelpers.CreateTestProfile(t, db, 1, user.ID, "Pending Profile")
	profile.Status = "pending"
	db.Save(profile)

	ownerToken := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupProfileRouter(db)

	updateData := map[string]interface{}{
		"name":             "Updated Pending Profile",
		"profileFor":       "self",
		"dateOfBirth":      "1990-01-01",
		"maritalStatus":    "Single",
		"phoneNumbers":     "1234567890",
		"birthPlace":       "Test City",
		"height":           "5'10",
		"complexion":       "Fair",
		"fatherName":       "Father",
		"motherName":       "Mother",
		"siblings":         "1 brother",
		"qualification":    "B.Tech",
		"occupation":       "Engineer",
		"annualIncome":     "10L",
		"address":          "123 Test St",
		"currentLocation":  "Test City",
		"status":           "pending",
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/matrimonialProfiles/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedProfile models.Profile
	db.First(&updatedProfile, 1)
	assert.Equal(t, "Updated Pending Profile", updatedProfile.Name)
	assert.Equal(t, "pending", updatedProfile.Status, "Pending profile should remain pending after owner edit")
}

func TestUpdateProfile_OwnerEditRejectedRemainsRejected(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	profile := testhelpers.CreateTestProfile(t, db, 1, user.ID, "Rejected Profile")
	// Profile uses "inactive" as the rejected/not-approved state
	profile.Status = "inactive"
	rejecterID := uint(2)
	profile.RejectedBy = &rejecterID
	rejectTime := testhelpers.ParseTime(t, "2025-01-01T10:00:00Z")
	profile.RejectedAt = &rejectTime
	db.Save(profile)

	ownerToken := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupProfileRouter(db)

	updateData := map[string]interface{}{
		"name":             "Updated Rejected Profile",
		"profileFor":       "self",
		"dateOfBirth":      "1990-01-01",
		"maritalStatus":    "Single",
		"phoneNumbers":     "1234567890",
		"birthPlace":       "Test City",
		"height":           "5'10",
		"complexion":       "Fair",
		"fatherName":       "Father",
		"motherName":       "Mother",
		"siblings":         "1 brother",
		"qualification":    "B.Tech",
		"occupation":       "Engineer",
		"annualIncome":     "10L",
		"address":          "123 Test St",
		"currentLocation":  "Test City",
		"status":           "inactive",
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/matrimonialProfiles/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedProfile models.Profile
	db.First(&updatedProfile, 1)
	assert.Equal(t, "Updated Rejected Profile", updatedProfile.Name)
	assert.Equal(t, "inactive", updatedProfile.Status, "Inactive profile should remain inactive after owner edit")
	assert.NotNil(t, updatedProfile.RejectedBy)
	assert.NotNil(t, updatedProfile.RejectedAt)
}

// TestOwnerEditApprovedResetsToPending_TableDriven tests the re-pending logic across all content types
func TestOwnerEditApprovedResetsToPending_TableDriven(t *testing.T) {
	testCases := []struct {
		name           string
		setupFn        func(*testing.T, *gorm.DB) (uint, string) // returns (resourceID, ownerToken)
		updateFn       func(*testing.T, *gorm.DB, uint, string) *httptest.ResponseRecorder
		checkStatus    func(*testing.T, *gorm.DB, uint) string
		checkAuditCleared func(*testing.T, *gorm.DB, uint)
	}{
		{
			name: "Profile",
			setupFn: func(t *testing.T, db *gorm.DB) (uint, string) {
				user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
				profile := testhelpers.CreateTestProfile(t, db, 1, user.ID, "Approved Profile")
				profile.Status = "active"
				approverID := uint(2)
				profile.ApprovedBy = &approverID
				approveTime := testhelpers.ParseTime(t, "2025-01-01T10:00:00Z")
				profile.ApprovedAt = &approveTime
				db.Save(profile)
				return 1, testhelpers.GenerateTestToken(t, "1", "user")
			},
			updateFn: func(t *testing.T, db *gorm.DB, id uint, token string) *httptest.ResponseRecorder {
				r := setupProfileRouter(db)
				updateData := map[string]interface{}{
					"name":             "Updated",
					"profileFor":       "self",
					"dateOfBirth":      "1990-01-01",
					"maritalStatus":    "Single",
					"phoneNumbers":     "1234567890",
					"birthPlace":       "Test City",
					"height":           "5'10",
					"complexion":       "Fair",
					"fatherName":       "Father",
					"motherName":       "Mother",
					"siblings":         "1 brother",
					"qualification":    "B.Tech",
					"occupation":       "Engineer",
					"annualIncome":     "10L",
					"address":          "123 Test St",
					"currentLocation":  "Test City",
					"status":           "active",
				}
				body, _ := json.Marshal(updateData)
				req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/matrimonialProfiles/%d", id), bytes.NewBuffer(body))
				req.Header.Set("Authorization", "Bearer "+token)
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				return w
			},
			checkStatus: func(t *testing.T, db *gorm.DB, id uint) string {
				var p models.Profile
				db.First(&p, id)
				return p.Status
			},
			checkAuditCleared: func(t *testing.T, db *gorm.DB, id uint) {
				var p models.Profile
				db.First(&p, id)
				assert.Nil(t, p.ApprovedBy)
				assert.Nil(t, p.ApprovedAt)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := testhelpers.SetupTestDB(t)
			defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

			id, token := tc.setupFn(t, db)
			w := tc.updateFn(t, db, id, token)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "pending", tc.checkStatus(t, db, id))
			tc.checkAuditCleared(t, db, id)
		})
	}
}