package controllers

import (
	initializers "SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Get all Businesses with pagination
func GetBusinesses(c *gin.Context) {
	var businesses []models.Business

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Apply public status filter - only show approved businesses to public
	query := utils.GetQueryWithPublicStatus(c, initializers.DB, "business")

	if err := query.Offset(offset).Limit(limit).Find(&businesses).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch businesses")
		return
	}

	var totalRecords int64
	if err := utils.GetQueryWithPublicStatus(c, initializers.DB.Model(&models.Business{}), "business").Count(&totalRecords).Error; err != nil {
		utils.RespondInternalError(c, "Failed to count businesses")
		return
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"page":         page,
		"limit":        limit,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
		"data":         businesses,
	})
}

// Get business by ID
func GetBusinessByID(c *gin.Context) {
	id := c.Param("id")
	var business models.Business

	// First, try to find the business without status filter to check ownership
	if err := initializers.DB.First(&business, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	// Check if user is admin
	userRole, hasRole := c.Get("role")
	if hasRole && userRole == "admin" {
		c.JSON(http.StatusOK, business)
		return
	}

	// For business, there's no direct owner field like profiles
	// Public users can only see Approved businesses
	if business.Status != "Approved" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	c.JSON(http.StatusOK, business)
}

// Create a new business
func CreateBusiness(c *gin.Context) {
	var business models.Business

	if err := c.ShouldBindJSON(&business); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Force Pending status - ignore any client-provided status
	business.Status = "Pending"
	business.CreatedAt = time.Now()
	business.UpdatedAt = time.Now()

	if err := initializers.DB.Create(&business).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusCreated, business)
}

// Update business
func UpdateBusiness(c *gin.Context) {
	id := c.Param("id")
	var business models.Business

	if err := initializers.DB.First(&business, id).Error; err != nil {
		utils.RespondNotFound(c, "Business not found")
		return
	}

	if err := c.ShouldBindJSON(&business); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	business.UpdatedAt = time.Now()
	if err := initializers.DB.Save(&business).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, business)
}

// Delete business
func DeleteBusiness(c *gin.Context) {
	id := c.Param("id")
	var business models.Business

	if err := initializers.DB.First(&business, id).Error; err != nil {
		utils.RespondNotFound(c, "Business not found")
		return
	}

	if err := initializers.DB.Delete(&business).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Business deleted successfully"})

	if err := initializers.DB.First(&business, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	initializers.DB.Delete(&business)
	c.JSON(http.StatusOK, gin.H{"message": "Business deleted successfully"})
}
