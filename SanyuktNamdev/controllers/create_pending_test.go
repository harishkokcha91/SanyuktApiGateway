package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"SanyuktNamdev/middleware"
	"SanyuktNamdev/models"
	"SanyuktNamdev/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBusiness_ForcesPendingStatus(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Business{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	userToken := testhelpers.GenerateTestToken(t, "1", "user")
	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	businessGroup := r.Group("/businesses")
	businessGroup.Use(middleware.AuthMiddleware())
	{
		businessGroup.POST("/", CreateBusiness)
	}

	testCases := []struct {
		name       string
		token      string
		body       string
		wantStatus string
	}{
		{
			name:       "User sends Approved - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Business","category":"Restaurant","owner":"Test Owner","email":"owner@test.com","location":"Test Location","city":"Test City","state":"Test State","country":"Test Country","status":"Approved"}`,
			wantStatus: "Pending",
		},
		{
			name:       "User sends Rejected - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Business","category":"Restaurant","owner":"Test Owner","email":"owner@test.com","location":"Test Location","city":"Test City","state":"Test State","country":"Test Country","status":"Rejected"}`,
			wantStatus: "Pending",
		},
		{
			name:       "User sends empty status - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Business","category":"Restaurant","owner":"Test Owner","email":"owner@test.com","location":"Test Location","city":"Test City","state":"Test State","country":"Test Country"}`,
			wantStatus: "Pending",
		},
		{
			name:       "Admin sends Approved - should still be Pending",
			token:      adminToken,
			body:       `{"name":"Test Business","category":"Restaurant","owner":"Test Owner","email":"owner@test.com","location":"Test Location","city":"Test City","state":"Test State","country":"Test Country","status":"Approved"}`,
			wantStatus: "Pending",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/businesses/", bytes.NewBufferString(tc.body))
			req.Header.Set("Authorization", "Bearer "+tc.token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var resp models.Business
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, tc.wantStatus, resp.Status, "Status should be %s, got %s", tc.wantStatus, resp.Status)
		})
	}
}

func TestCreateEvent_ForcesPendingStatus(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Event{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	userToken := testhelpers.GenerateTestToken(t, "1", "user")
	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	eventGroup := r.Group("/events")
	eventGroup.Use(middleware.AuthMiddleware())
	{
		eventGroup.POST("/", CreateEvent)
	}

	testCases := []struct {
		name       string
		token      string
		body       string
		wantStatus string
	}{
		{
			name:       "User sends Approved - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Event","event_date":"2025-12-31","venue":"Test Venue","city":"Test City","state":"Test State","country":"Test Country","organizer":"Test Organizer","email":"organizer@test.com","category":"Tech","status":"Upcoming"}`,
			wantStatus: "Pending",
		},
		{
			name:       "User sends Completed - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Event","event_date":"2025-12-31","venue":"Test Venue","city":"Test City","state":"Test State","country":"Test Country","organizer":"Test Organizer","email":"organizer@test.com","category":"Tech","status":"Completed"}`,
			wantStatus: "Pending",
		},
		{
			name:       "User sends empty status - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Event","event_date":"2025-12-31","venue":"Test Venue","city":"Test City","state":"Test State","country":"Test Country","organizer":"Test Organizer","email":"organizer@test.com","category":"Tech"}`,
			wantStatus: "Pending",
		},
		{
			name:       "Admin sends Upcoming - should still be Pending",
			token:      adminToken,
			body:       `{"name":"Test Event","event_date":"2025-12-31","venue":"Test Venue","city":"Test City","state":"Test State","country":"Test Country","organizer":"Test Organizer","email":"organizer@test.com","category":"Tech","status":"Upcoming"}`,
			wantStatus: "Pending",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/events/", bytes.NewBufferString(tc.body))
			req.Header.Set("Authorization", "Bearer "+tc.token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var resp models.Event
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, tc.wantStatus, resp.Status, "Status should be %s, got %s", tc.wantStatus, resp.Status)
		})
	}
}

func TestCreateAchievement_ForcesPendingStatus(t *testing.T) {
	db := testhelpers.SetupTestDBWithModels(t, &models.User{}, &models.Achievement{})
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	userToken := testhelpers.GenerateTestToken(t, "1", "user")
	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	achievementGroup := r.Group("/achievements")
	achievementGroup.Use(middleware.AuthMiddleware())
	{
		achievementGroup.POST("/", CreateAchievement)
	}

	testCases := []struct {
		name       string
		token      string
		body       string
		wantStatus string
	}{
		{
			name:       "User sends Approved - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Achievement","achievement_type":"Academic","achievement":"Test Achievement Detail","date_of_achievement":"2025-01-01","status":"Approved"}`,
			wantStatus: "Pending",
		},
		{
			name:       "User sends Rejected - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Achievement","achievement_type":"Academic","achievement":"Test Achievement Detail","date_of_achievement":"2025-01-01","status":"Rejected"}`,
			wantStatus: "Pending",
		},
		{
			name:       "User sends empty status - should be Pending",
			token:      userToken,
			body:       `{"name":"Test Achievement","achievement_type":"Academic","achievement":"Test Achievement Detail","date_of_achievement":"2025-01-01"}`,
			wantStatus: "Pending",
		},
		{
			name:       "Admin sends Approved - should still be Pending",
			token:      adminToken,
			body:       `{"name":"Test Achievement","achievement_type":"Academic","achievement":"Test Achievement Detail","date_of_achievement":"2025-01-01","status":"Approved"}`,
			wantStatus: "Pending",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/achievements/", bytes.NewBufferString(tc.body))
			req.Header.Set("Authorization", "Bearer "+tc.token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var resp models.Achievement
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, tc.wantStatus, resp.Status, "Status should be %s, got %s", tc.wantStatus, resp.Status)
		})
	}
}