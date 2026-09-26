package controllers

import (
	"SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	var users []models.User

	// Get the page and limit from the query parameters
	page := c.DefaultQuery("page", "1")    // Default page to 1 if not provided
	limit := c.DefaultQuery("limit", "10") // Default limit to 10 if not provided

	// Convert page and limit to integers
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

	// Calculate the offset
	offset := (pageInt - 1) * limitInt

	// Fetch users from the database with pagination
	if err := database.DB.Offset(offset).Limit(limitInt).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Get the total number of records
	var totalRecords int64
	if err := database.DB.Model(&models.User{}).Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch total records"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(limitInt)))

	// Return paginated data
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
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}
	c.JSON(http.StatusOK, user)
}

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	if err := database.DB.Create(&user).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	if err := database.DB.Save(&user).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully", "user": user})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.User{}, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
