package controllers

import (
	"SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/notifications"
	"SanyuktNamdev/utils"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AdminApprovalType represents the valid content types for admin approval
type AdminApprovalType string

const (
	ApprovalTypeProfile     AdminApprovalType = "profile"
	ApprovalTypeBusiness    AdminApprovalType = "business"
	ApprovalTypeEvent       AdminApprovalType = "event"
	ApprovalTypeAchievement AdminApprovalType = "achievement"
)

// AdminApprovalRequest represents the request body for approval/rejection
type AdminApprovalRequest struct {
	Reason string `json:"reason,omitempty"`
}

// getModelFromType returns the appropriate model instance based on type
func getModelFromType(approvalType AdminApprovalType) (interface{}, *gorm.DB, error) {
	db := database.DB
	switch approvalType {
	case ApprovalTypeProfile:
		return &models.Profile{}, db.Model(&models.Profile{}), nil
	case ApprovalTypeBusiness:
		return &models.Business{}, db.Model(&models.Business{}), nil
	case ApprovalTypeEvent:
		return &models.Event{}, db.Model(&models.Event{}), nil
	case ApprovalTypeAchievement:
		return &models.Achievement{}, db.Model(&models.Achievement{}), nil
	default:
		return nil, nil, fmt.Errorf("invalid type: must be one of profile, business, event, achievement")
	}
}

// getAuditFieldsFromType returns the audit field names for the given type
func getAuditFieldsFromType(approvalType AdminApprovalType) (approvedByField, approvedAtField, rejectedByField, rejectedAtField string) {
	switch approvalType {
	case ApprovalTypeProfile:
		return "approved_by", "approved_at", "rejected_by", "rejected_at"
	case ApprovalTypeBusiness:
		return "approved_by", "approved_at", "rejected_by", "rejected_at"
	case ApprovalTypeEvent:
		return "approved_by", "approved_at", "rejected_by", "rejected_at"
	case ApprovalTypeAchievement:
		return "approved_by", "approved_at", "rejected_by", "rejected_at"
	default:
		return "", "", "", ""
	}
}

// getOwnerID extracts the owner/user ID from the model instance
func getOwnerID(modelInstance interface{}, approvalType AdminApprovalType) (uint, error) {
	modelValue := reflect.ValueOf(modelInstance).Elem()

	var userIDField string
	switch approvalType {
	case ApprovalTypeProfile:
		userIDField = "UserId"
	case ApprovalTypeBusiness:
		// Business doesn't have a direct owner field, use 0 for now
		return 0, nil
	case ApprovalTypeEvent:
		// Event doesn't have a direct owner field, use 0 for now
		return 0, nil
	case ApprovalTypeAchievement:
		// Achievement doesn't have a direct owner field, use 0 for now
		return 0, nil
	default:
		return 0, fmt.Errorf("unknown approval type")
	}

	userIDFieldVal := modelValue.FieldByName(userIDField)
	if userIDFieldVal.IsValid() && userIDFieldVal.CanInterface() {
		return userIDFieldVal.Interface().(uint), nil
	}
	return 0, nil
}

// getContentName extracts a display name from the model instance
func getContentName(modelInstance interface{}, approvalType AdminApprovalType) string {
	modelValue := reflect.ValueOf(modelInstance).Elem()

	var nameField string
	switch approvalType {
	case ApprovalTypeProfile:
		nameField = "Name"
	case ApprovalTypeBusiness:
		nameField = "Name"
	case ApprovalTypeEvent:
		nameField = "Name"
	case ApprovalTypeAchievement:
		nameField = "Name"
	default:
		return "Record"
	}

	nameFieldVal := modelValue.FieldByName(nameField)
	if nameFieldVal.IsValid() && nameFieldVal.CanInterface() {
		return nameFieldVal.Interface().(string)
	}
	return "Record"
}

// getTypeDisplayName returns the display name for the approval type
func getTypeDisplayName(approvalType AdminApprovalType) string {
	switch approvalType {
	case ApprovalTypeProfile:
		return "Profile"
	case ApprovalTypeBusiness:
		return "Business"
	case ApprovalTypeEvent:
		return "Event"
	case ApprovalTypeAchievement:
		return "Achievement"
	default:
		return "Record"
	}
}

// sendApprovalNotification sends approval notification to the content owner
func sendApprovalNotification(approvalType AdminApprovalType, modelInstance interface{}) {
	ownerID, err := getOwnerID(modelInstance, approvalType)
	if err != nil || ownerID == 0 {
		return // No owner to notify (e.g., Business, Event, Achievement don't have owner field yet)
	}

	contentName := getContentName(modelInstance, approvalType)
	typeDisplayName := getTypeDisplayName(approvalType)

	notifService := notifications.NewService()
	if err := notifService.NotifyOnApproval(ownerID, typeDisplayName, contentName); err != nil {
		// Log error but don't fail the approval
		fmt.Printf("Failed to send approval notification: %v\n", err)
	}
}

// sendRejectionNotification sends rejection notification to the content owner
func sendRejectionNotification(approvalType AdminApprovalType, modelInstance interface{}, reason string) {
	ownerID, err := getOwnerID(modelInstance, approvalType)
	if err != nil || ownerID == 0 {
		return // No owner to notify
	}

	contentName := getContentName(modelInstance, approvalType)
	typeDisplayName := getTypeDisplayName(approvalType)

	notifService := notifications.NewService()
	if err := notifService.NotifyOnRejection(ownerID, typeDisplayName, contentName, reason); err != nil {
		fmt.Printf("Failed to send rejection notification: %v\n", err)
	}
}

// AdminApprove handles PATCH /admin/:type/:id/approve
func AdminApprove(c *gin.Context) {
	// Verify admin role
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		utils.RespondForbidden(c, "Admin access required")
		return
	}

	// Get type and ID from URL
	approvalType := AdminApprovalType(strings.ToLower(c.Param("type")))
	id := c.Param("id")

	// Validate type
	modelInstance, modelDB, err := getModelFromType(approvalType)
	if err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Find the record
	if err := modelDB.First(modelInstance, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondNotFound(c, "Record not found")
			return
		}
		utils.RespondDBError(c, err)
		return
	}

	// Get admin user ID
	adminUserID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	adminIDStr := fmt.Sprintf("%v", adminUserID)
	adminID, err := strconv.ParseUint(adminIDStr, 10, 64)
	if err != nil {
		utils.RespondInternalError(c, "Invalid user ID")
		return
	}

	// Update status and audit fields
	now := time.Now()
	_, _, _, _ = getAuditFieldsFromType(approvalType)

	// Use reflection to set fields on the model instance
	modelValue := reflect.ValueOf(modelInstance).Elem()
	statusField := modelValue.FieldByName("Status")
	approvedByFieldVal := modelValue.FieldByName("ApprovedBy")
	approvedAtFieldVal := modelValue.FieldByName("ApprovedAt")

	if statusField.IsValid() && statusField.CanSet() {
		statusField.SetString("Approved")
	}
	if approvedByFieldVal.IsValid() && approvedByFieldVal.CanSet() {
		adminIDVal := uint(adminID)
		approvedByFieldVal.Set(reflect.ValueOf(&adminIDVal))
	}
	if approvedAtFieldVal.IsValid() && approvedAtFieldVal.CanSet() {
		approvedAtFieldVal.Set(reflect.ValueOf(&now))
	}

	if err := database.DB.Save(modelInstance).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	// Send approval notification to content owner
	sendApprovalNotification(approvalType, modelInstance)

	// Fetch updated record
	if err := database.DB.First(modelInstance, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Record approved successfully",
		"data":    modelInstance,
	})
}

// AdminReject handles PATCH /admin/:type/:id/reject
func AdminReject(c *gin.Context) {
	// Verify admin role
	role, exists := c.Get("role")
	if !exists || role != "admin" {
		utils.RespondForbidden(c, "Admin access required")
		return
	}

	// Get type and ID from URL
	approvalType := AdminApprovalType(strings.ToLower(c.Param("type")))
	id := c.Param("id")

	// Validate type
	modelInstance, modelDB, err := getModelFromType(approvalType)
	if err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Find the record
	if err := modelDB.First(modelInstance, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondNotFound(c, "Record not found")
			return
		}
		utils.RespondDBError(c, err)
		return
	}

	// Get admin user ID
	adminUserID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	adminIDStr := fmt.Sprintf("%v", adminUserID)
	adminID, err := strconv.ParseUint(adminIDStr, 10, 64)
	if err != nil {
		utils.RespondInternalError(c, "Invalid user ID")
		return
	}

	// Parse request body for optional reason
	var req AdminApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Reason is optional, ignore binding errors
		req.Reason = ""
	}

	// Update status and audit fields
	now := time.Now()
	_, _, _, _ = getAuditFieldsFromType(approvalType)

	// Use reflection to set fields on the model instance
	modelValue := reflect.ValueOf(modelInstance).Elem()
	statusField := modelValue.FieldByName("Status")
	rejectedByFieldVal := modelValue.FieldByName("RejectedBy")
	rejectedAtFieldVal := modelValue.FieldByName("RejectedAt")

	if statusField.IsValid() && statusField.CanSet() {
		statusField.SetString("Rejected")
	}
	if rejectedByFieldVal.IsValid() && rejectedByFieldVal.CanSet() {
		adminIDVal := uint(adminID)
		rejectedByFieldVal.Set(reflect.ValueOf(&adminIDVal))
	}
	if rejectedAtFieldVal.IsValid() && rejectedAtFieldVal.CanSet() {
		rejectedAtFieldVal.Set(reflect.ValueOf(&now))
	}

	if err := database.DB.Save(modelInstance).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	// Send rejection notification to content owner
	sendRejectionNotification(approvalType, modelInstance, req.Reason)

	// Fetch updated record
	if err := modelDB.First(modelInstance, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Record rejected successfully",
		"data":    modelInstance,
	})
}