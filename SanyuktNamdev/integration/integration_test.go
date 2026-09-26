package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"SanyuktNamdev/database"
	"SanyuktNamdev/middleware"
	"SanyuktNamdev/models"
	"SanyuktNamdev/routes"
	"SanyuktNamdev/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestIntegration_LoginProtectedFlow tests the full login -> protected route flow
// This test requires a running PostgreSQL instance.
// Set TEST_DATABASE_URL env var or use docker-compose test profile.
func TestIntegration_LoginProtectedFlow(t *testing.T) {
	// Skip if no test database configured
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:admin@localhost:6501/sanyukt_test?sslmode=disable"
		t.Logf("Using default test DB: %s", dbURL)
	}

	// Connect to test database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-at-least-32-bytes-long")
	os.Setenv("ENV", "test")
	utils.SetTestJWTSecret("test-secret-key-that-is-at-least-32-bytes-long")

	// Run migrations on test DB (create tables if not exist)
	err = db.AutoMigrate(&models.User{}, &models.Profile{}, &models.Event{}, &models.Business{}, &models.Achievement{})
	require.NoError(t, err)

	// Replace package-level DB
	database.DB = db

	// Clean up test data
	db.Exec("DELETE FROM users WHERE email LIKE 'integration_test_%'")

	// Create test router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.Logger(middleware.NewLogger("test", false)))
	r.Use(middleware.SecurityCORSMiddleware(middleware.DefaultCORSConfig()))
	r.Use(middleware.SecurityHeaders(false))
	r.Use(middleware.BodyLimitMiddleware())
	r.Use(middleware.GlobalRateLimiter())

	// Health endpoints
	r.GET("/healthz", func(c *gin.Context) {
		ctx := c.Request.Context()
		sqlDB, err := database.WithContext(ctx).DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Setup all routes
	routes.SetupAuthRoutes(r)
	routes.UserRoutes(r)
	routes.ProfileRoutes(r)
	routes.AchievementRoutes(r)
	routes.BusinessRoutes(r)
	routes.EventRoutes(r)

	// Test user credentials
	testEmail := "integration_test_" + time.Now().Format("150405") + "@test.com"
	testPassword := "TestPassword123!"
	testName := "Integration Test User"

	// 1. Register user
	t.Run("Register", func(t *testing.T) {
		registerBody := map[string]interface{}{
			"email":         testEmail,
			"password":      testPassword,
			"name":          testName,
			"age":           25,
			"phoneNumbers":  "1234567890",
			"status":        "active",
		}
		bodyBytes, _ := json.Marshal(registerBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code, "Registration should succeed: %s", w.Body.String())

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp["id"])
	})

	// 2. Login
	var authToken string
	t.Run("Login", func(t *testing.T) {
		loginBody := map[string]string{
			"email":    testEmail,
			"password": testPassword,
		}
		bodyBytes, _ := json.Marshal(loginBody)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Login should succeed: %s", w.Body.String())

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		token, ok := resp["token"].(string)
		require.True(t, ok, "Token should be in response")
		assert.NotEmpty(t, token)
		authToken = token
	})

	// 3. Access protected route with token
	t.Run("ProtectedRoute_Authenticated", func(t *testing.T) {
		require.NotEmpty(t, authToken, "Need auth token from login")

		req := httptest.NewRequest(http.MethodGet, "/auth/protected", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Protected route should be accessible with valid token: %s", w.Body.String())

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Contains(t, resp, "message")
	})

	// 4. Access protected route without token
	t.Run("ProtectedRoute_NoAuth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "No token provided")
	})

	// 5. Access protected route with invalid token
	t.Run("ProtectedRoute_InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	// 6. Test health endpoint
	t.Run("HealthEndpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Health check should pass: %s", w.Body.String())

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "healthy", resp["status"])
	})

	// 7. Test ready endpoint
	t.Run("ReadyEndpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "ready", resp["status"])
	})

	// 8. Test ownership enforcement on user profile
	t.Run("UserProfile_Ownership", func(t *testing.T) {
		require.NotEmpty(t, authToken)

		// Get user profile
		req := httptest.NewRequest(http.MethodGet, "/matrimonialProfiles/byuserId/1", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should work (or return empty if no profiles)
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
	})

	// 9. Test admin-only endpoint rejection for regular user
	t.Run("AdminEndpoint_RejectedForUser", func(t *testing.T) {
		require.NotEmpty(t, authToken)

		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBufferString(`{
			"name": "Test Event",
			"event_date": "2025-12-31",
			"venue": "Venue",
			"city": "City",
			"state": "State",
			"country": "Country",
			"organizer": "Org",
			"email": "org@test.com",
			"category": "Tech"
		}`))
		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code, "Regular user should be forbidden from admin endpoints")
		assert.Contains(t, w.Body.String(), "Insufficient permissions")
	})
}

// TestIntegration_HealthEndpoint tests health endpoint with database connectivity
func TestIntegration_HealthEndpoint(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err)

	database.DB = db

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	r.GET("/healthz", func(c *gin.Context) {
		ctx := c.Request.Context()
		sqlDB, err := database.WithContext(ctx).DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Test healthy
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "healthy", resp["status"])
}

// TestIntegration_RateLimiting tests that rate limiting works
func TestIntegration_RateLimiting(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err)
	database.DB = db

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.AuthRateLimiter()) // Stricter rate limiting for auth
	r.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Make requests up to limit (5 per minute for auth)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"test@test.com","password":"pass"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i+1)
	}

	// 6th request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"test@test.com","password":"pass"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "6th request should be rate limited")
	assert.Contains(t, w.Body.String(), "Too many attempts")
}

// TestIntegration_CORS tests CORS headers
func TestIntegration_CORS(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err)
	database.DB = db

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.SecurityCORSMiddleware(middleware.DefaultCORSConfig()))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

// TestIntegration_SecurityHeaders tests security headers are present
func TestIntegration_SecurityHeaders(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err)
	database.DB = db

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.SecurityHeaders(false))

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check security headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

// TestIntegration_RequestID tests request ID propagation
func TestIntegration_RequestID(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err)
	database.DB = db

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())

	r.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get("request_id")
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Test with custom request ID
	customID := "custom-request-id-123"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", customID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, customID, w.Header().Get("X-Request-ID"))

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, customID, resp["request_id"])

	// Test auto-generated request ID
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	assert.NotEmpty(t, w2.Header().Get("X-Request-ID"))
	assert.NotEqual(t, customID, w2.Header().Get("X-Request-ID"))
}

// TestIntegration_FullUserLifecycle tests complete user lifecycle
func TestIntegration_FullUserLifecycle(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{}, &models.Profile{})
	require.NoError(t, err)

	database.DB = db
	os.Setenv("JWT_SECRET", "test-secret-key-that-is-at-least-32-bytes-long")
	utils.SetTestJWTSecret("test-secret-key-that-is-at-least-32-bytes-long")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.AuthMiddleware())
	r.Use(middleware.BodyLimitMiddleware())

	routes.UserRoutes(r)
	routes.ProfileRoutes(r)

	// Register
	email := "lifecycle_test_" + time.Now().Format("150405") + "@test.com"
	registerBody := map[string]interface{}{
		"email":         email,
		"password":      "Password123!",
		"name":          "Lifecycle User",
		"age":           30,
		"phoneNumbers":  "1234567890",
		"status":        "active",
	}
	bodyBytes, _ := json.Marshal(registerBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var registerResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResp)
	userID := registerResp["id"]

	// Login
	loginBody := map[string]string{"email": email, "password": "Password123!"}
	bodyBytes, _ = json.Marshal(loginBody)

	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	token := loginResp["token"].(string)

	// Get own user
	req = httptest.NewRequest(http.MethodGet, "/users/"+string(rune(userID.(float64))), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update own user
	updateBody := map[string]interface{}{"name": "Updated Name"}
	bodyBytes, _ = json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPut, "/users/"+string(rune(userID.(float64))), bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Create profile
	profileBody := map[string]interface{}{
		"profileFor":      "self",
		"name":            "My Profile",
		"dateOfBirth":     "1990-01-01",
		"maritalStatus":   "Single",
		"phoneNumbers":    "1234567890",
		"birthPlace":      "City",
		"height":          "5'10",
		"complexion":      "Fair",
		"fatherName":      "Father",
		"motherName":      "Mother",
		"siblings":        "1",
		"qualification":   "B.Tech",
		"occupation":      "Engineer",
		"annualIncome":    "10L",
		"address":         "123 St",
		"currentLocation": "City",
	}
	bodyBytes, _ = json.Marshal(profileBody)
	req = httptest.NewRequest(http.MethodPost, "/matrimonialProfiles", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Delete profile
	req = httptest.NewRequest(http.MethodDelete, "/matrimonialProfiles/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete user
	req = httptest.NewRequest(http.MethodDelete, "/users/"+string(rune(userID.(float64))), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}