package controllers

import (
	"bytes"
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

func setupAdminTestDB(t *testing.T) *gorm.DB {
	return testhelpers.SetupTestDBWithModels(t, &models.Event{}, &models.Business{}, &models.Achievement{}, &models.User{}, &models.Profile{})
}

func setupAdminRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	// Event routes
	eventGroup := r.Group("/events")
	{
		eventGroup.GET("/", GetEvents)
		eventGroup.GET("/:id", GetEventByID)

		admin := eventGroup.Group("")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
		{
			admin.POST("/", CreateEvent)
			admin.PUT("/:id", UpdateEvent)
			admin.DELETE("/:id", DeleteEvent)
		}
	}

	// Business routes
	businessGroup := r.Group("/businesses")
	{
		businessGroup.GET("/", GetBusinesses)
		businessGroup.GET("/:id", GetBusinessByID)

		admin := businessGroup.Group("")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
		{
			admin.POST("/", CreateBusiness)
			admin.PUT("/:id", UpdateBusiness)
			admin.DELETE("/:id", DeleteBusiness)
		}
	}

	// Achievement routes
	achievementGroup := r.Group("/achievements")
	{
		achievementGroup.GET("/", GetAchievements)
		achievementGroup.GET("/:id", GetAchievementByID)

		admin := achievementGroup.Group("")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
		{
			admin.POST("/", CreateAchievement)
			admin.PUT("/:id", UpdateAchievement)
			admin.DELETE("/:id", DeleteAchievement)
		}
	}

	// Admin approval routes
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		adminGroup.PATCH("/:type/:id/approve", AdminApprove)
		adminGroup.PATCH("/:type/:id/reject", AdminReject)
	}

	return r
}

// TestEventAdminOnlyEnforcement tests admin-only write access for events
func TestEventAdminOnlyEnforcement(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestEvent(t, db, 1)
	userToken := testhelpers.GenerateTestToken(t, "1", "user")
	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	r := setupAdminRouter(db)

	testCases := []struct {
		name       string
		method     string
		path       string
		token      string
		body       string
		expectCode int
	}{
		// Public read - both should work
		{"GET events - user", http.MethodGet, "/events/", userToken, "", http.StatusOK},
		{"GET events - admin", http.MethodGet, "/events/", adminToken, "", http.StatusOK},
		{"GET event by ID - user", http.MethodGet, "/events/1", userToken, "", http.StatusOK},
		{"GET event by ID - admin", http.MethodGet, "/events/1", adminToken, "", http.StatusOK},

		// Admin write - user should be forbidden
		{"POST event - user", http.MethodPost, "/events/", userToken, `{"name":"Test Event","event_date":"2025-12-31","venue":"Test Venue","city":"Test City","state":"Test State","country":"Test Country","organizer":"Test Organizer","email":"organizer@test.com","category":"Tech","status":"upcoming"}`, http.StatusForbidden},
		{"POST event - admin", http.MethodPost, "/events/", adminToken, `{"name":"Test Event","event_date":"2025-12-31","venue":"Test Venue","city":"Test City","state":"Test State","country":"Test Country","organizer":"Test Organizer","email":"organizer@test.com","category":"Tech","status":"upcoming"}`, http.StatusCreated},
		{"PUT event - user", http.MethodPut, "/events/1", userToken, `{"name":"Updated Event"}`, http.StatusForbidden},
		{"PUT event - admin", http.MethodPut, "/events/1", adminToken, `{"name":"Updated Event"}`, http.StatusOK},
		{"DELETE event - user", http.MethodDelete, "/events/1", userToken, "", http.StatusForbidden},
		{"DELETE event - admin", http.MethodDelete, "/events/1", adminToken, "", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tc.body != "" {
				body = bytes.NewBufferString(tc.body)
			} else {
				body = bytes.NewBufferString("")
			}

			req := httptest.NewRequest(tc.method, tc.path, body)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectCode, w.Code, "test case: %s", tc.name)
		})
	}
}

// TestBusinessAdminOnlyEnforcement tests admin-only write access for businesses
func TestBusinessAdminOnlyEnforcement(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestBusiness(t, db, 1)
	userToken := testhelpers.GenerateTestToken(t, "1", "user")
	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	r := setupAdminRouter(db)

	testCases := []struct {
		name       string
		method     string
		path       string
		token      string
		body       string
		expectCode int
	}{
		// Public read
		{"GET businesses - user", http.MethodGet, "/businesses/", userToken, "", http.StatusOK},
		{"GET businesses - admin", http.MethodGet, "/businesses/", adminToken, "", http.StatusOK},
		{"GET business by ID - user", http.MethodGet, "/businesses/1", userToken, "", http.StatusOK},
		{"GET business by ID - admin", http.MethodGet, "/businesses/1", adminToken, "", http.StatusOK},

		// Admin write
		{"POST business - user", http.MethodPost, "/businesses/", userToken, `{"name":"Test Business","category":"Restaurant","owner":"Test Owner","email":"owner@test.com","location":"Test Location","city":"Test City","state":"Test State","country":"Test Country"}`, http.StatusForbidden},
		{"POST business - admin", http.MethodPost, "/businesses/", adminToken, `{"name":"Test Business","category":"Restaurant","owner":"Test Owner","email":"owner@test.com","location":"Test Location","city":"Test City","state":"Test State","country":"Test Country"}`, http.StatusCreated},
		{"PUT business - user", http.MethodPut, "/businesses/1", userToken, `{"name":"Updated Business"}`, http.StatusForbidden},
		{"PUT business - admin", http.MethodPut, "/businesses/1", adminToken, `{"name":"Updated Business"}`, http.StatusOK},
		{"DELETE business - user", http.MethodDelete, "/businesses/1", userToken, "", http.StatusForbidden},
		{"DELETE business - admin", http.MethodDelete, "/businesses/1", adminToken, "", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tc.body != "" {
				body = bytes.NewBufferString(tc.body)
			} else {
				body = bytes.NewBufferString("")
			}

			req := httptest.NewRequest(tc.method, tc.path, body)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectCode, w.Code, "test case: %s", tc.name)
		})
	}
}

// TestAchievementAdminOnlyEnforcement tests admin-only write access for achievements
func TestAchievementAdminOnlyEnforcement(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestAchievement(t, db, 1)
	userToken := testhelpers.GenerateTestToken(t, "1", "user")
	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	r := setupAdminRouter(db)

	testCases := []struct {
		name       string
		method     string
		path       string
		token      string
		body       string
		expectCode int
	}{
		// Public read
		{"GET achievements - user", http.MethodGet, "/achievements/", userToken, "", http.StatusOK},
		{"GET achievements - admin", http.MethodGet, "/achievements/", adminToken, "", http.StatusOK},
		{"GET achievement by ID - user", http.MethodGet, "/achievements/1", userToken, "", http.StatusOK},
		{"GET achievement by ID - admin", http.MethodGet, "/achievements/1", adminToken, "", http.StatusOK},

		// Admin write
		{"POST achievement - user", http.MethodPost, "/achievements/", userToken, `{"name":"Test Achievement","achievement_type":"Academic","achievement":"Test Achievement Detail","date_of_achievement":"2025-01-01","status":"Pending"}`, http.StatusForbidden},
		{"POST achievement - admin", http.MethodPost, "/achievements/", adminToken, `{"name":"Test Achievement","achievement_type":"Academic","achievement":"Test Achievement Detail","date_of_achievement":"2025-01-01","status":"Pending"}`, http.StatusCreated},
		{"PUT achievement - user", http.MethodPut, "/achievements/1", userToken, `{"name":"Updated Achievement"}`, http.StatusForbidden},
		{"PUT achievement - admin", http.MethodPut, "/achievements/1", adminToken, `{"name":"Updated Achievement"}`, http.StatusOK},
		{"DELETE achievement - user", http.MethodDelete, "/achievements/1", userToken, "", http.StatusForbidden},
		{"DELETE achievement - admin", http.MethodDelete, "/achievements/1", adminToken, "", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tc.body != "" {
				body = bytes.NewBufferString(tc.body)
			} else {
				body = bytes.NewBufferString("")
			}

			req := httptest.NewRequest(tc.method, tc.path, body)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectCode, w.Code, "test case: %s", tc.name)
		})
	}
}

// TestAdminEndpoints_NoAuth tests that unauthenticated requests are rejected
func TestAdminEndpoints_NoAuth(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	r := setupAdminRouter(db)

	testCases := []struct {
		name   string
		method string
		path   string
	}{
		{"POST events", http.MethodPost, "/events/"},
		{"PUT events", http.MethodPut, "/events/1"},
		{"DELETE events", http.MethodDelete, "/events/1"},
		{"POST businesses", http.MethodPost, "/businesses/"},
		{"PUT businesses", http.MethodPut, "/businesses/1"},
		{"DELETE businesses", http.MethodDelete, "/businesses/1"},
		{"POST achievements", http.MethodPost, "/achievements/"},
		{"PUT achievements", http.MethodPut, "/achievements/1"},
		{"DELETE achievements", http.MethodDelete, "/achievements/1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString("{}"))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "test case: %s", tc.name)
			assert.Contains(t, w.Body.String(), "No token provided")
		})
	}
}

// TestAdminEndpoints_InvalidToken tests that invalid tokens are rejected
func TestAdminEndpoints_InvalidToken(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	r := setupAdminRouter(db)

	req := httptest.NewRequest(http.MethodPost, "/events/", bytes.NewBufferString("{}"))
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

// Table-driven test for admin enforcement across all resources
func TestAdminEnforcement_TableDriven(t *testing.T) {
	testCases := []struct {
		resource   string
		createFunc func(*testing.T, *gorm.DB) uint
		listPath   string
		detailPath string
		createBody string
		updateBody string
	}{
		{
			resource: "events",
			createFunc: func(t *testing.T, db *gorm.DB) uint {
				return testhelpers.CreateTestEvent(t, db, 1).ID
			},
			listPath:   "/events/",
			detailPath: "/events/1",
			createBody: `{"name":"Test Event","event_date":"2025-12-31","venue":"Test Venue","city":"Test City","state":"Test State","country":"Test Country","organizer":"Test Organizer","email":"organizer@test.com","category":"Tech","status":"upcoming"}`,
			updateBody: `{"name":"Updated Event"}`,
		},
		{
			resource: "businesses",
			createFunc: func(t *testing.T, db *gorm.DB) uint {
				return testhelpers.CreateTestBusiness(t, db, 1).ID
			},
			listPath:   "/businesses/",
			detailPath: "/businesses/1",
			createBody: `{"name":"Test Business","category":"Restaurant","owner":"Test Owner","email":"owner@test.com","location":"Test Location","city":"Test City","state":"Test State","country":"Test Country"}`,
			updateBody: `{"name":"Updated Business"}`,
		},
		{
			resource: "achievements",
			createFunc: func(t *testing.T, db *gorm.DB) uint {
				return testhelpers.CreateTestAchievement(t, db, 1).ID
			},
			listPath:   "/achievements/",
			detailPath: "/achievements/1",
			createBody: `{"name":"Test Achievement","achievement_type":"Academic","achievement":"Test Achievement Detail","date_of_achievement":"2025-01-01","status":"Pending"}`,
			updateBody: `{"name":"Updated Achievement"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.resource, func(t *testing.T) {
			db := setupAdminTestDB(t)
			defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

			tc.createFunc(t, db)
			userToken := testhelpers.GenerateTestToken(t, "1", "user")
			adminToken := testhelpers.GenerateTestToken(t, "2", "admin")
			r := setupAdminRouter(db)

			// Test read access for both
			for _, path := range []string{tc.listPath, tc.detailPath} {
				for _, tt := range []struct {
					name  string
					token string
					code  int
				}{
					{"user read", userToken, http.StatusOK},
					{"admin read", adminToken, http.StatusOK},
				} {
					t.Run(tt.name+" "+path, func(t *testing.T) {
						req := httptest.NewRequest(http.MethodGet, path, nil)
						req.Header.Set("Authorization", "Bearer "+tt.token)
						w := httptest.NewRecorder()
						r.ServeHTTP(w, req)
						assert.Equal(t, tt.code, w.Code)
					})
				}
			}

			// Test write access - user forbidden, admin allowed
			writeTests := []struct {
				name   string
				method string
				path   string
				body   string
				userCode int
				adminCode int
			}{
				{"create", http.MethodPost, tc.listPath, tc.createBody, http.StatusForbidden, http.StatusCreated},
				{"update", http.MethodPut, tc.detailPath, tc.updateBody, http.StatusForbidden, http.StatusOK},
				{"delete", http.MethodDelete, tc.detailPath, "", http.StatusForbidden, http.StatusOK},
			}

			for _, wt := range writeTests {
				t.Run(wt.name, func(t *testing.T) {
					// User attempt
					req1 := httptest.NewRequest(wt.method, wt.path, bytes.NewBufferString(wt.body))
					req1.Header.Set("Authorization", "Bearer "+userToken)
					req1.Header.Set("Content-Type", "application/json")
					w1 := httptest.NewRecorder()
					r.ServeHTTP(w1, req1)
					assert.Equal(t, wt.userCode, w1.Code, "user should be forbidden")

					// Admin attempt
					req2 := httptest.NewRequest(wt.method, wt.path, bytes.NewBufferString(wt.body))
					req2.Header.Set("Authorization", "Bearer "+adminToken)
					req2.Header.Set("Content-Type", "application/json")
					w2 := httptest.NewRecorder()
					r.ServeHTTP(w2, req2)
					assert.Equal(t, wt.adminCode, w2.Code, "admin should be allowed")
				})
			}
		})
	}
}