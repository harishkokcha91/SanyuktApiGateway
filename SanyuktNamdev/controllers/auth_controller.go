package controllers

import (
	"SanyuktNamdev/config"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register user
func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	// Hash password before storing
	hashedPassword, _ := utils.HashPassword(user.Password)
	user.Password = hashedPassword

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": user})
}

// Login user
func Login(c *gin.Context) {
	var input models.User
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if user exists
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Compare password
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	// Generate JWT Token
	userIDStr := fmt.Sprintf("%d", user.ID)
	token, _ := utils.GenerateToken(userIDStr)
	// claims, err := utils.ValidateToken(token)
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
	// 	c.Abort()
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user, "message": "Login successful"})
}

// Register user
func RegisterUserIfExistReturnUser(c *gin.Context) {
	var user models.User
	// Bind the JSON payload to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Check if the user already exists by email or phone number
	var existingUser models.User
	// if err := config.DB.Where("email = ? OR phone_numbers = ?", user.Email, user.PhoneNumbers).First(&existingUser).Error; err == nil {
	if err := config.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		// User already exists, return existing user details
		c.JSON(http.StatusOK, gin.H{"message": "User already exists", "user": existingUser})
		return
	}

	// Hash password before storing
	hashedPassword, _ := utils.HashPassword(user.Password)
	user.Password = hashedPassword

	// Create the new user in the database
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to register user"})
		return
	}

	// Return success response for newly registered user
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": user})
}
