package controllers

import (
	initializers "brilliant-student/database"
	"brilliant-student/models"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Create Achievement
func CreateAchievement(c *gin.Context) {
	var achievement models.Achievement
	if err := c.ShouldBindJSON(&achievement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	initializers.DB.Create(&achievement)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}

	// Calculate offset for pagination
	offset := (pageInt - 1) * limitInt

	// Fetch paginated achievements
	if err := initializers.DB.Offset(offset).Limit(limitInt).Find(&achievements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch achievements"})
		return
	}

	// Get the total number of records
	var totalRecords int64
	if err := initializers.DB.Model(&models.Achievement{}).Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch total records"})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Achievement not found"})
		return
	}
	c.JSON(http.StatusOK, achievement)
}

// Update Achievement
func UpdateAchievement(c *gin.Context) {
	id := c.Param("id")
	var achievement models.Achievement
	if err := initializers.DB.First(&achievement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Achievement not found"})
		return
	}

	if err := c.ShouldBindJSON(&achievement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	initializers.DB.Save(&achievement)
	c.JSON(http.StatusOK, achievement)
}

// Delete Achievement
func DeleteAchievement(c *gin.Context) {
	id := c.Param("id")
	if err := initializers.DB.Delete(&models.Achievement{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Achievement not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Achievement deleted successfully"})
}
