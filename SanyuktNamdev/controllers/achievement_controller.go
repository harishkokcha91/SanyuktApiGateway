package controllers

import (
	initializers "SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Create Achievement
func CreateAchievement(c *gin.Context) {
	var achievement models.Achievement
	if err := c.ShouldBindJSON(&achievement); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Force Pending status - ignore any client-provided status
	achievement.Status = "Pending"

	if err := initializers.DB.Create(&achievement).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusCreated, achievement)
}

func GetAchievements(c *gin.Context) {
	var achievements []models.Achievement

	// Get page and limit from query parameters (default: page=1, limit=10)
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// Convert page & limit to integers
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		utils.RespondValidationError(c, "Invalid page number")
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		utils.RespondValidationError(c, "Invalid limit")
		return
	}

	// Calculate offset for pagination
	offset := (pageInt - 1) * limitInt

	// Apply public status filter - only show approved achievements to public
	query := utils.GetQueryWithPublicStatus(c, initializers.DB, "achievement")

	// Fetch paginated achievements
	if err := query.Offset(offset).Limit(limitInt).Find(&achievements).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch achievements")
		return
	}

	// Get the total number of records
	var totalRecords int64
	if err := utils.GetQueryWithPublicStatus(c, initializers.DB.Model(&models.Achievement{}), "achievement").Count(&totalRecords).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch total records")
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(limitInt)))

	// Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"page":         pageInt,
		"limit":        limitInt,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
		"data":         achievements,
	})
}

// Get Achievement by ID
func GetAchievementByID(c *gin.Context) {
	id := c.Param("id")
	var achievement models.Achievement
	if err := initializers.DB.First(&achievement, id).Error; err != nil {
		utils.RespondNotFound(c, "Achievement not found")
		return
	}

	// Check if user is admin
	userRole, hasRole := c.Get("role")
	if hasRole && userRole == "admin" {
		c.JSON(http.StatusOK, achievement)
		return
	}

	// Public users can only see Approved achievements
	if achievement.Status != "Approved" {
		utils.RespondNotFound(c, "Achievement not found")
		return
	}
	c.JSON(http.StatusOK, achievement)
}

// Update Achievement
func UpdateAchievement(c *gin.Context) {
	id := c.Param("id")
	var achievement models.Achievement
	if err := initializers.DB.First(&achievement, id).Error; err != nil {
		utils.RespondNotFound(c, "Achievement not found")
		return
	}

	// Get current user info from context
	userRole, hasUserRole := c.Get("role")

	// Ownership check: Achievement doesn't have a direct owner field
	// For now, only admins can update achievements
	isAdmin := hasUserRole && userRole == "admin"
	if !isAdmin {
		utils.RespondForbidden(c, "Admin access required to update achievement")
		return
	}

	// Bind JSON to existing achievement struct (preserves existing values for fields not in JSON)
	if err := c.ShouldBindJSON(&achievement); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Owner edit on Approved record -> reset to Pending
	// Achievement uses "Approved" as approved status, "Pending" as pending status
	// Note: Since only admins can update, this logic applies if ownership is added later
	isApproved := achievement.Status == "Approved"

	// Prevent non-admins from directly setting status to Approved/Rejected
	// (Currently only admins can update, so this is for future ownership implementation)
	if !isAdmin {
		if achievement.Status == "Approved" || achievement.Status == "Rejected" {
			// This would need the original status to compare, but since only admins update,
			// we can't easily implement this without fetching original status first.
			// For now, we rely on the fact that only admins can update.
		}
	}

	// Preserve audit fields if status is being reset by owner
	// (Not applicable for admin-only updates, but kept for future ownership implementation)
	if !isAdmin && isApproved {
		achievement.Status = "Pending"
		achievement.ApprovedBy = nil
		achievement.ApprovedAt = nil
		// Preserve RejectedBy/RejectedAt if previously rejected
	}

	if err := initializers.DB.Save(&achievement).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, achievement)
}

// Delete Achievement
func DeleteAchievement(c *gin.Context) {
	id := c.Param("id")
	if err := initializers.DB.Delete(&models.Achievement{}, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Achievement deleted successfully"})
}
