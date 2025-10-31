package controllers

import (
	initializers "SanyuktNamdev/database"
	"SanyuktNamdev/models"
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

	if err := initializers.DB.Offset(offset).Limit(limit).Find(&businesses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch businesses"})
		return
	}

	var totalRecords int64
	initializers.DB.Model(&models.Business{}).Count(&totalRecords)

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

	if err := initializers.DB.First(&business, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	c.JSON(http.StatusOK, business)
}

// Create a new business
func CreateBusiness(c *gin.Context) {
	var business models.Business

	if err := c.ShouldBindJSON(&business); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	business.CreatedAt = time.Now()
	business.UpdatedAt = time.Now()

	initializers.DB.Create(&business)
	c.JSON(http.StatusCreated, business)
}

// Update business
func UpdateBusiness(c *gin.Context) {
	id := c.Param("id")
	var business models.Business

	if err := initializers.DB.First(&business, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	if err := c.ShouldBindJSON(&business); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	business.UpdatedAt = time.Now()
	initializers.DB.Save(&business)

	c.JSON(http.StatusOK, business)
}

// Delete business
func DeleteBusiness(c *gin.Context) {
	id := c.Param("id")
	var business models.Business

	if err := initializers.DB.First(&business, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	initializers.DB.Delete(&business)
	c.JSON(http.StatusOK, gin.H{"message": "Business deleted successfully"})
}
