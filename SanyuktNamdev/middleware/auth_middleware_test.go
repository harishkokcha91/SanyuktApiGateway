package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"SanyuktNamdev/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a gin router with the middleware under test
func setupTestRouter(middlewareFn gin.HandlerFunc, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middlewareFn)
	r.GET("/test", handler)
	return r
}

// testHandler is a simple handler for testing
func testHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// generateTestToken generates a test token
func generateTestToken(t *testing.T, userID string, role string) string {
	// Set test JWT secret
	utils.SetTestJWTSecret("test-secret-key-that-is-at-least-32-bytes-long")
	token, err := utils.GenerateToken(userID, role)
	require.NoError(t, err)
	return token
}

// TestAuthMiddleware_ValidToken tests that a valid token passes through
func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Generate a valid token
	token := generateTestToken(t, "123", "user")

	// Setup router with AuthMiddleware
	r := setupTestRouter(AuthMiddleware(), testHandler)

	// Make request with valid token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestAuthMiddleware_MissingToken tests that missing token returns 401
func TestAuthMiddleware_MissingToken(t *testing.T) {
	r := setupTestRouter(AuthMiddleware(), testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "No token provided")
}

// TestAuthMiddleware_MalformedToken tests that malformed token returns 401
func TestAuthMiddleware_MalformedToken(t *testing.T) {
	testCases := []struct {
		name           string
		authHeader     string
		expectedError  string
	}{
		{
			name:           "empty bearer",
			authHeader:     "Bearer ",
			expectedError:  "Invalid Authorization header format",
		},
		{
			name:           "no bearer prefix",
			authHeader:     "InvalidToken123",
			expectedError:  "Invalid Authorization header format",
		},
		{
			name:           "wrong prefix",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedError:  "Invalid Authorization header format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := setupTestRouter(AuthMiddleware(), testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.authHeader)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "test case: %s", tc.name)
			assert.Contains(t, w.Body.String(), tc.expectedError, "test case: %s", tc.name)
		})
	}
}

// TestAuthMiddleware_InvalidToken tests that invalid/expired token returns 401
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	testCases := []struct {
		name          string
		token         string
		expectedError string
	}{
		{
			name:          "random string",
			token:         "invalid.token.string",
			expectedError: "Invalid token",
		},
		{
			name:          "tampered token",
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiIxMjMiLCJyb2xlIjoidXNlciIsImV4cCI6MTk5OTk5OTk5OX0.invalid_signature",
			expectedError: "Invalid token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := setupTestRouter(AuthMiddleware(), testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "test case: %s", tc.name)
			assert.Contains(t, w.Body.String(), tc.expectedError, "test case: %s", tc.name)
		})
	}
}

// TestRequireRole_ValidRole tests that correct role passes through
func TestRequireRole_ValidRole(t *testing.T) {
	// Generate admin token
	token := generateTestToken(t, "123", "admin")

	// Setup router with both middlewares
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(), RequireRole("admin"))
	r.GET("/admin", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

// TestRequireRole_WrongRole tests that wrong role returns 403
func TestRequireRole_WrongRole(t *testing.T) {
	// Generate user token (not admin)
	token := generateTestToken(t, "123", "user")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(), RequireRole("admin"))
	r.GET("/admin", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
}

// TestRequireRole_NoRoleInContext tests missing role returns 403
func TestRequireRole_NoRoleInContext(t *testing.T) {
	// Create a token without role claim
	token := generateTestToken(t, "123", "")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(), RequireRole("admin"))
	r.GET("/admin", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
}

// TestRequireRole_MissingAuth tests that missing auth returns 401 before role check
func TestRequireRole_MissingAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(), RequireRole("admin"))
	r.GET("/admin", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "No token provided")
}

// TestAuthMiddleware_SetsContextValues tests that userid and role are set in context
func TestAuthMiddleware_SetsContextValues(t *testing.T) {
	token := generateTestToken(t, "456", "admin")

	var capturedUserID, capturedRole interface{}
	captureHandler := func(c *gin.Context) {
		capturedUserID, _ = c.Get("userid")
		capturedRole, _ = c.Get("role")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/test", captureHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "456", capturedUserID)
	assert.Equal(t, "admin", capturedRole)
}

// TestAuthMiddleware_DefaultRoleUser tests that missing role defaults to "user"
func TestAuthMiddleware_DefaultRoleUser(t *testing.T) {
	// Generate token with empty role
	token := generateTestToken(t, "789", "")

	var capturedRole interface{}
	captureHandler := func(c *gin.Context) {
		capturedRole, _ = c.Get("role")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/test", captureHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user", capturedRole)
}

// TestAuthMiddleware_TokenWithExtraWhitespace tests bearer token with extra spaces
func TestAuthMiddleware_TokenWithExtraWhitespace(t *testing.T) {
	token := generateTestToken(t, "123", "user")

	testCases := []struct {
		name       string
		authHeader string
	}{
		{"extra spaces", "Bearer  " + token},
		{"tab", "Bearer\t" + token},
		{"newline", "Bearer\n" + token},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := setupTestRouter(AuthMiddleware(), testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.authHeader)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Should fail because of extra whitespace
			assert.Equal(t, http.StatusUnauthorized, w.Code, "test case: %s", tc.name)
		})
	}
}

// TestAuthMiddleware_CaseInsensitiveBearer tests Bearer case variations
func TestAuthMiddleware_CaseInsensitiveBearer(t *testing.T) {
	token := generateTestToken(t, "123", "user")

	testCases := []struct {
		name       string
		authHeader string
	}{
		{"lowercase", "bearer " + token},
		{"uppercase", "BEARER " + token},
		{"mixed", "BeArEr " + token},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := setupTestRouter(AuthMiddleware(), testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.authHeader)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code, "test case: %s", tc.name)
			assert.Contains(t, w.Body.String(), "Invalid Authorization header format")
		})
	}
}

// TestRequireRole_MultipleRoles tests different role combinations
func TestRequireRole_MultipleRoles(t *testing.T) {
	testCases := []struct {
		name           string
		userRole       string
		requiredRole   string
		expectedStatus int
	}{
		{"user accessing user", "user", "user", http.StatusOK},
		{"admin accessing user", "admin", "user", http.StatusOK}, // admin can access user endpoints
		{"user accessing admin", "user", "admin", http.StatusForbidden},
		{"admin accessing admin", "admin", "admin", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token := generateTestToken(t, "123", tc.userRole)

			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(AuthMiddleware(), RequireRole(tc.requiredRole))
			r.GET("/test", testHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code, "test case: %s", tc.name)
		})
	}
}

// TestAuthMiddleware_ExpiredToken tests expired token handling
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// This would require generating a token with past expiration
	// Since GenerateToken always creates future-dated tokens,
	// we test by manually creating an expired token
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiIxMjMiLCJyb2xlIjoidXNlciIsImV4cCI6MSwiaWF0IjowfQ.invalid"

	r := setupTestRouter(AuthMiddleware(), testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")
}

// TestAuthMiddleware_SQLInjectionInToken tests that token with SQL injection attempt is rejected
func TestAuthMiddleware_SQLInjectionInToken(t *testing.T) {
	// Token with SQL injection attempt in claims (would fail signature validation anyway)
	maliciousToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyaWQiOiIxMjM7IERST1AgVEFCTEUgdXNlcnM7LS0iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE5OTk5OTk5OTl9.invalid"

	r := setupTestRouter(AuthMiddleware(), testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", maliciousToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// BenchmarkAuthMiddleware benchmarks the auth middleware
func BenchmarkAuthMiddleware(b *testing.B) {
	utils.SetTestJWTSecret("test-secret-key-that-is-at-least-32-bytes-long")
	token, _ := utils.GenerateToken("123", "user")
	r := setupTestRouter(AuthMiddleware(), testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// BenchmarkRequireRole benchmarks the role middleware
func BenchmarkRequireRole(b *testing.B) {
	utils.SetTestJWTSecret("test-secret-key-that-is-at-least-32-bytes-long")
	token, _ := utils.GenerateToken("123", "admin")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(), RequireRole("admin"))
	r.GET("/test", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}

// TestAuthMiddleware_ConcurrentRequests tests middleware under concurrent load
func TestAuthMiddleware_ConcurrentRequests(t *testing.T) {
	token := generateTestToken(t, "123", "user")

	r := setupTestRouter(AuthMiddleware(), testHandler)

	const concurrency = 100
	done := make(chan int, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			done <- w.Code
		}()
	}

	for i := 0; i < concurrency; i++ {
		code := <-done
		assert.Equal(t, http.StatusOK, code)
	}
}

// TestAuthMiddleware_ContextIsolation tests that context values don't leak between requests
func TestAuthMiddleware_ContextIsolation(t *testing.T) {
	userToken := generateTestToken(t, "111", "user")
	adminToken := generateTestToken(t, "222", "admin")

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware())
	r.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("userid")
		role, _ := c.Get("role")
		c.JSON(http.StatusOK, gin.H{"userid": userID, "role": role})
	})

	// Make user request
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("Authorization", "Bearer "+userToken)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)
	assert.Contains(t, w1.Body.String(), `"userid":"111"`)
	assert.Contains(t, w1.Body.String(), `"role":"user"`)

	// Make admin request
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Contains(t, w2.Body.String(), `"userid":"222"`)
	assert.Contains(t, w2.Body.String(), `"role":"admin"`)
}