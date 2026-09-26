package utils

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PublicStatusFilter represents the public-visible status for each content type
type PublicStatusFilter string

const (
	ProfilePublicStatus    PublicStatusFilter = "active"
	BusinessPublicStatus   PublicStatusFilter = "Approved"
	EventPublicStatus      PublicStatusFilter = "Approved"
	AchievementPublicStatus PublicStatusFilter = "Approved"
)

// GetPublicStatus returns the public-visible status for a given model type
func GetPublicStatus(modelType string) string {
	switch modelType {
	case "profile":
		return string(ProfilePublicStatus)
	case "business":
		return string(BusinessPublicStatus)
	case "event":
		return string(EventPublicStatus)
	case "achievement":
		return string(AchievementPublicStatus)
	default:
		return ""
	}
}

// ApplyPublicStatusFilter applies the public status filter to a query
// If the user is authenticated and is the owner or admin, it returns all statuses
// Otherwise, it filters to only public-visible status
func ApplyPublicStatusFilter(c *gin.Context, db *gorm.DB, modelType string) *gorm.DB {
	// Check if user is authenticated
	_, hasUserID := c.Get("userid")
	userRole, hasUserRole := c.Get("role")

	if !hasUserID {
		// Not authenticated - only public status
		return db.Where("status = ?", GetPublicStatus(modelType))
	}

	// Admin can see everything
	if hasUserRole && userRole == "admin" {
		return db
	}

	// For owner-specific endpoints, we need to check ownership
	// This is handled at the controller level by adding owner filter
	// Here we just return the base query - controller will add ownership check
	return db.Where("status = ?", GetPublicStatus(modelType))
}

// ApplyOwnerOrPublicFilter applies status filter for owner-accessible endpoints
// Owner can see their own items regardless of status
// Others only see public status
func ApplyOwnerOrPublicFilter(c *gin.Context, db *gorm.DB, modelType string, ownerField string, ownerID uint) *gorm.DB {
	userID, hasUserID := c.Get("userid")
	userRole, hasUserRole := c.Get("role")

	// Admin can see everything
	if hasUserRole && userRole == "admin" {
		return db
	}

	// Check if requesting user is the owner
	if hasUserID {
		reqUserIDStr := fmt.Sprintf("%v", userID)
		reqUserID, _ := strconv.ParseUint(reqUserIDStr, 10, 64)
		if uint(reqUserID) == ownerID {
			// Owner can see all their items regardless of status
			return db
		}
	}

	// Not owner or not authenticated - only public status
	return db.Where("status = ?", GetPublicStatus(modelType))
}

// GetQueryWithPublicStatus is a generic helper for list endpoints
func GetQueryWithPublicStatus(c *gin.Context, db *gorm.DB, modelType string) *gorm.DB {
	return ApplyPublicStatusFilter(c, db, modelType)
}