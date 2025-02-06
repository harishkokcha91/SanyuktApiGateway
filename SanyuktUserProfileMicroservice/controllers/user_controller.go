package controllers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
	"userprofile-service/database"
	"userprofile-service/models"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	var users []models.Profile

	// Get the page and limit from the query parameters
	page := c.DefaultQuery("page", "1")    // Default page to 1 if not provided
	limit := c.DefaultQuery("limit", "10") // Default limit to 10 if not provided

	// Convert page and limit to integers
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

	// Calculate the offset based on the page number
	offset := (pageInt - 1) * limitInt

	// Fetch the data from the database with pagination
	if err := database.DB.Offset(offset).Limit(limitInt).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Get the total number of records
	var totalRecords int64
	if err := database.DB.Model(&models.Profile{}).Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch total records"})
		return
	}

	// Calculate totalPages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(limitInt)))

	// Return the paginated data as response
	c.JSON(http.StatusOK, gin.H{
		"page":         pageInt,
		"limit":        limitInt,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
		"data":         users,
	})
}

func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	var user models.Profile
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func GetProfileByUserID(c *gin.Context) {
	fmt.Println("GetProfileByUserID")
	id := c.Param("id") // Get user ID from the URL
	var users []models.Profile

	// Find the user in the database
	if err := database.DB.Find(&users, "user_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user models.Profile
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Create(&user)
	c.JSON(http.StatusCreated, user)
}

func convertInDateFormate(inputDate string) time.Time {
	const customDateFormat = "2006-01-02" // Format for date without time component

	// Parse the date from JSON
	dateOfBirth, err := time.Parse(customDateFormat, inputDate)
	if err != nil {
		log.Fatal("Error parsing date: ", err)
	}
	return dateOfBirth
}
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.Profile
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Save(&user)
	c.JSON(http.StatusOK, user)
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Profile{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
