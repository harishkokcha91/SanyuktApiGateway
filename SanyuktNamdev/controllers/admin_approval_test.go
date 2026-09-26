package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"SanyuktNamdev/models"
	"SanyuktNamdev/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminApprove_Success(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create test records
	profile := testhelpers.CreateTestProfile(t, db, 1, 1, "Test Profile")
	business := testhelpers.CreateTestBusiness(t, db, 1)
	event := testhelpers.CreateTestEvent(t, db, 1)
	achievement := testhelpers.CreateTestAchievement(t, db, 1)

	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	r := setupAdminRouter(db)

	testCases := []struct {
		name         string
		approvalType string
		id           uint
	}{
		{"Approve Profile", "profile", profile.ID},
		{"Approve Business", "business", business.ID},
		{"Approve Event", "event", event.ID},
		{"Approve Achievement", "achievement", achievement.ID},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("/admin/%s/%d/approve", tc.approvalType, tc.id)
			req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBufferString("{}"))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "test case: %s", tc.name)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, "Record approved successfully", resp["message"])
			assert.NotNil(t, resp["data"])
		})
	}

	// Verify audit fields were populated
	var p models.Profile
	db.First(&p, profile.ID)
	assert.Equal(t, "Approved", p.Status)
	assert.NotNil(t, p.ApprovedBy)
	assert.NotNil(t, p.ApprovedAt)
	assert.Equal(t, uint(2), *p.ApprovedBy)

	var b models.Business
	db.First(&b, business.ID)
	assert.Equal(t, "Approved", b.Status)
	assert.NotNil(t, b.ApprovedBy)

	var e models.Event
	db.First(&e, event.ID)
	assert.Equal(t, "Approved", e.Status)
	assert.NotNil(t, e.ApprovedBy)

	var a models.Achievement
	db.First(&a, achievement.ID)
	assert.Equal(t, "Approved", a.Status)
	assert.NotNil(t, a.ApprovedBy)
}

func TestAdminReject_Success(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	// Create test records
	profile := testhelpers.CreateTestProfile(t, db, 1, 1, "Test Profile")
	business := testhelpers.CreateTestBusiness(t, db, 1)
	event := testhelpers.CreateTestEvent(t, db, 1)
	achievement := testhelpers.CreateTestAchievement(t, db, 1)

	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	r := setupAdminRouter(db)

	testCases := []struct {
		name         string
		approvalType string
		id           uint
	}{
		{"Reject Profile", "profile", profile.ID},
		{"Reject Business", "business", business.ID},
		{"Reject Event", "event", event.ID},
		{"Reject Achievement", "achievement", achievement.ID},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body := `{"reason": "Inappropriate content"}`
			url := fmt.Sprintf("/admin/%s/%d/reject", tc.approvalType, tc.id)
			req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBufferString(body))
			req.Header.Set("Authorization", "Bearer "+adminToken)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "test case: %s", tc.name)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, "Record rejected successfully", resp["message"])
			assert.NotNil(t, resp["data"])
		})
	}

	// Verify audit fields were populated
	var p models.Profile
	db.First(&p, profile.ID)
	assert.Equal(t, "Rejected", p.Status)
	assert.NotNil(t, p.RejectedBy)
	assert.NotNil(t, p.RejectedAt)
	assert.Equal(t, uint(2), *p.RejectedBy)
}

func TestAdminApprove_NonAdminBlocked(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	profile := testhelpers.CreateTestProfile(t, db, 1, 1, "Test Profile")
	userToken := testhelpers.GenerateTestToken(t, "1", "user")

	r := setupAdminRouter(db)

	url := fmt.Sprintf("/admin/profile/%d/approve", profile.ID)
	req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBufferString("{}"))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminApprove_InvalidType(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	r := setupAdminRouter(db)

	req := httptest.NewRequest(http.MethodPatch, "/admin/invalid/1/approve", bytes.NewBufferString("{}"))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAdminApprove_NotFound(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	r := setupAdminRouter(db)

	req := httptest.NewRequest(http.MethodPatch, "/admin/profile/99999/approve", bytes.NewBufferString("{}"))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAdminApprove_MissingAuth(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	r := setupAdminRouter(db)

	req := httptest.NewRequest(http.MethodPatch, "/admin/profile/1/approve", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAdminReject_WithReason(t *testing.T) {
	db := setupAdminTestDB(t)
	defer func() { sqlDB, _ := db.DB(); sqlDB.Close() }()

	profile := testhelpers.CreateTestProfile(t, db, 1, 1, "Test Profile")
	adminToken := testhelpers.GenerateTestToken(t, "2", "admin")

	r := setupAdminRouter(db)

	body := `{"reason": "Spam content"}`
	url := fmt.Sprintf("/admin/profile/%d/reject", profile.ID)
	req := httptest.NewRequest(http.MethodPatch, url, bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}