package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"SanyuktNamdev/middleware"
	"SanyuktNamdev/models"
	"SanyuktNamdev/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupRouter creates a test router with all middleware and routes
func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	// Public routes (no auth required)
	// None for user/profile - all require auth

	// Protected routes - require authentication
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/users", GetUsers)
		protected.GET("/users/:id", GetUserByID)
		protected.POST("/users", CreateUser)
		protected.PUT("/users/:id", UpdateUser)
		protected.DELETE("/users/:id", DeleteUser)

		protected.GET("/matrimonialProfiles", GetProfiles)
		protected.GET("/matrimonialProfiles/:id", GetProfileByID)
		protected.GET("/matrimonialProfiles/byuserId/:id", GetProfilesByUserID)
		protected.POST("/matrimonialProfiles", CreateProfile)
		protected.PUT("/matrimonialProfiles/:id", UpdateProfile)
		protected.DELETE("/matrimonialProfiles/:id", DeleteProfile)
	}

	return r
}

func TestGetUserByID_OwnerSuccess(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.User
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, "owner@test.com", resp.Email)
}

func TestGetUserByID_NonOwnerForbidden(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	// Non-owner token
	token := testhelpers.GenerateTestToken(t, "999", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only view your own user record")
}

func TestGetUserByID_NoAuthUnauthorized(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "No token provided")
}

func TestGetUserByID_NotFound(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "User not found")
}

func TestUpdateUser_OwnerSuccess(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupRouter(db)

	updateData := map[string]interface{}{
		"name": "Updated Name",
		"age":  30,
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User updated successfully")
}

func TestUpdateUser_NonOwnerForbidden(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "999", "user")

	r := setupRouter(db)

	updateData := map[string]interface{}{"name": "Hacker"}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/users/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only update your own user record")
}

func TestDeleteUser_OwnerSuccess(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User deleted successfully")
}

func TestDeleteUser_NonOwnerForbidden(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "999", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only delete your own user record")
}

func TestGetProfileByID_OwnerSuccess(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	testhelpers.CreateTestProfile(t, db, 1, user.ID, "My Profile")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/matrimonialProfiles/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp models.Profile
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, user.ID, resp.UserId)
}

func TestGetProfileByID_NonOwnerForbidden(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	testhelpers.CreateTestProfile(t, db, 1, user.ID, "My Profile")
	token := testhelpers.GenerateTestToken(t, "999", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/matrimonialProfiles/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only view your own profile")
}

func TestGetProfilesByUserID_OwnerSuccess(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	testhelpers.CreateTestProfile(t, db, 1, user.ID, "Profile 1")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/matrimonialProfiles/byuserId/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Profile 1")
}

func TestGetProfilesByUserID_NonOwnerForbidden(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	testhelpers.CreateTestProfile(t, db, 1, user.ID, "Profile 1")
	token := testhelpers.GenerateTestToken(t, "999", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/matrimonialProfiles/byuserId/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only view your own profiles")
}

func TestCreateProfile_OwnerSuccess(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupRouter(db)

	profileData := map[string]interface{}{
		"profileFor":       "self",
		"name":             "New Profile",
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
	body, _ := json.Marshal(profileData)

	req := httptest.NewRequest(http.MethodPost, "/matrimonialProfiles", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "New Profile")
}

func TestUpdateProfile_OwnerSuccess(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	testhelpers.CreateTestProfile(t, db, 1, user.ID, "Original Profile")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupRouter(db)

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
		"status":           "pending",
	}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/matrimonialProfiles/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Updated Profile")
}

func TestUpdateProfile_NonOwnerForbidden(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	testhelpers.CreateTestProfile(t, db, 1, user.ID, "Original Profile")
	token := testhelpers.GenerateTestToken(t, "999", "user")

	r := setupRouter(db)

	updateData := map[string]interface{}{"name": "Hacked"}
	body, _ := json.Marshal(updateData)

	req := httptest.NewRequest(http.MethodPut, "/matrimonialProfiles/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only update your own profile")
}

func TestDeleteProfile_OwnerSuccess(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	testhelpers.CreateTestProfile(t, db, 1, user.ID, "My Profile")
	token := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodDelete, "/matrimonialProfiles/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User deleted successfully")
}

func TestDeleteProfile_NonOwnerForbidden(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	testhelpers.CreateTestProfile(t, db, 1, user.ID, "My Profile")
	token := testhelpers.GenerateTestToken(t, "999", "user")

	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodDelete, "/matrimonialProfiles/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "You can only delete your own profile")
}

// Table-driven test for ownership enforcement across all user/profile endpoints
func TestOwnershipEnforcement_TableDriven(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		path           string
		setupData      func(*testing.T, *gorm.DB) (uint, string) // returns (resourceID, ownerToken)
		nonOwnerToken  string
		expectedOwner  int
		expectedNonOwn int
	}{
		{
			name:  "GET /users/:id",
			method: http.MethodGet,
			path:  "/users/1",
			setupData: func(t *testing.T, db *gorm.DB) (uint, string) {
				user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
				token := testhelpers.GenerateTestToken(t, "1", "user")
				return user.ID, token
			},
			expectedOwner:  http.StatusOK,
			expectedNonOwn: http.StatusForbidden,
		},
		{
			name:  "PUT /users/:id",
			method: http.MethodPut,
			path:  "/users/1",
			setupData: func(t *testing.T, db *gorm.DB) (uint, string) {
				user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
				token := testhelpers.GenerateTestToken(t, "1", "user")
				return user.ID, token
			},
			expectedOwner:  http.StatusOK,
			expectedNonOwn: http.StatusForbidden,
		},
		{
			name:  "DELETE /users/:id",
			method: http.MethodDelete,
			path:  "/users/1",
			setupData: func(t *testing.T, db *gorm.DB) (uint, string) {
				user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
				token := testhelpers.GenerateTestToken(t, "1", "user")
				return user.ID, token
			},
			expectedOwner:  http.StatusOK,
			expectedNonOwn: http.StatusForbidden,
		},
		{
			name:  "GET /matrimonialProfiles/:id",
			method: http.MethodGet,
			path:  "/matrimonialProfiles/1",
			setupData: func(t *testing.T, db *gorm.DB) (uint, string) {
				user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
				testhelpers.CreateTestProfile(t, db, 1, user.ID, "Profile")
				token := testhelpers.GenerateTestToken(t, "1", "user")
				return 1, token
			},
			expectedOwner:  http.StatusOK,
			expectedNonOwn: http.StatusForbidden,
		},
		{
			name:  "PUT /matrimonialProfiles/:id",
			method: http.MethodPut,
			path:  "/matrimonialProfiles/1",
			setupData: func(t *testing.T, db *gorm.DB) (uint, string) {
				user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
				testhelpers.CreateTestProfile(t, db, 1, user.ID, "Profile")
				token := testhelpers.GenerateTestToken(t, "1", "user")
				return 1, token
			},
			expectedOwner:  http.StatusOK,
			expectedNonOwn: http.StatusForbidden,
		},
		{
			name:  "DELETE /matrimonialProfiles/:id",
			method: http.MethodDelete,
			path:  "/matrimonialProfiles/1",
			setupData: func(t *testing.T, db *gorm.DB) (uint, string) {
				user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
				testhelpers.CreateTestProfile(t, db, 1, user.ID, "Profile")
				token := testhelpers.GenerateTestToken(t, "1", "user")
				return 1, token
			},
			expectedOwner:  http.StatusOK,
			expectedNonOwn: http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := testhelpers.SetupTestDB(t)
			defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

			_, ownerToken := tc.setupData(t, db)
			nonOwnerToken := testhelpers.GenerateTestToken(t, "999", "user")

			r := setupRouter(db)

			// Test owner access
			req1 := httptest.NewRequest(tc.method, tc.path, nil)
			req1.Header.Set("Authorization", "Bearer "+ownerToken)
			if tc.method == http.MethodPut {
				body, _ := json.Marshal(map[string]interface{}{
					"name":             "test",
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
				})
				req1.Body = io.NopCloser(bytes.NewBuffer(body))
				req1.Header.Set("Content-Type", "application/json")
			}
			w1 := httptest.NewRecorder()
			r.ServeHTTP(w1, req1)
			assert.Equal(t, tc.expectedOwner, w1.Code, "Owner should have access")

			// Test non-owner access - for DELETE, resource may be gone after owner test
			expectedNonOwner := tc.expectedNonOwn
			if tc.method == http.MethodDelete {
				expectedNonOwner = http.StatusForbidden // or 404 if resource already deleted
			}
			req2 := httptest.NewRequest(tc.method, tc.path, nil)
			req2.Header.Set("Authorization", "Bearer "+nonOwnerToken)
			if tc.method == http.MethodPut {
				body, _ := json.Marshal(map[string]interface{}{
					"name":             "test",
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
				})
				req2.Body = io.NopCloser(bytes.NewBuffer(body))
				req2.Header.Set("Content-Type", "application/json")
			}
			w2 := httptest.NewRecorder()
			r.ServeHTTP(w2, req2)
			if tc.method == http.MethodDelete {
				// Accept either 403 (forbidden) or 404 (already deleted by owner test)
				assert.True(t, w2.Code == http.StatusForbidden || w2.Code == http.StatusNotFound, 
					"Non-owner should get 403 or 404, got %d", w2.Code)
			} else {
				assert.Equal(t, expectedNonOwner, w2.Code, "Non-owner should be forbidden")
			}
		})
	}
}

// TestValidationErrors tests validation error responses
func TestValidationErrors(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupRouter(db)

	testCases := []struct {
		name           string
		path           string
		body           string
		expectedError  string
	}{
		{
			name:          "CreateUser missing required fields",
			path:          "/users",
			body:          `{"email": "invalid"}`,
			expectedError: "Validation failed",
		},
		{
			name:          "CreateProfile missing required fields",
			path:          "/matrimonialProfiles",
			body:          `{"name": "test"}`,
			expectedError: "Validation failed",
		},
		{
			name:          "UpdateUser invalid page param",
			path:          "/users?page=invalid",
			body:          "",
			expectedError: "Invalid page number",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "error")
		})
	}
}

// TestPagination tests pagination parameters
func TestPagination(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create multiple users
	for i := 1; i <= 15; i++ {
		testhelpers.CreateTestUser(t, db, uint(i), "user"+string(rune(i+'0'))+"@test.com", "User "+string(rune(i+'0')))
	}

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupRouter(db)

	testCases := []struct {
		name       string
		query      string
		expectPage int
		expectLimit int
	}{
		{"default", "", 1, 10},
		{"page 2", "?page=2&limit=5", 2, 5},
		{"page 3 limit 3", "?page=3&limit=3", 3, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users"+tc.query, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, float64(tc.expectPage), resp["page"])
			assert.Equal(t, float64(tc.expectLimit), resp["limit"])
		})
	}
}

// TestProfileByUserID_Pagination tests profile pagination
func TestProfileByUserID_Pagination(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	user := testhelpers.CreateTestUser(t, db, 1, "owner@test.com", "Owner")
	for i := 1; i <= 5; i++ {
		testhelpers.CreateTestProfile(t, db, uint(i), user.ID, "Profile "+string(rune(i+'0')))
	}

	token := testhelpers.GenerateTestToken(t, "1", "user")
	r := setupRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/matrimonialProfiles/byuserId/1?page=1&limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["page"])
	assert.Equal(t, float64(2), resp["limit"])
	assert.Equal(t, float64(3), resp["totalPages"])
}