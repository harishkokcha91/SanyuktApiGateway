package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"userprofile-service/database"
	"userprofile-service/models"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {
	var users []models.UserProfile
	database.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

func GetUserByID(c *gin.Context) {
	id := c.Param("id")
	var user models.UserProfile
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func GetProfileByUserID(c *gin.Context) {
	fmt.Println("GetProfileByUserID")
	id := c.Param("id") // Get user ID from the URL
	var users []models.UserProfile

	// Find the user in the database
	if err := database.DB.Find(&users, "user_id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user models.UserProfile
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
	var user models.UserProfile
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
	if err := database.DB.Delete(&models.UserProfile{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
