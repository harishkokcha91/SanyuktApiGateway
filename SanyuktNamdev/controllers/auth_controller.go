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
		utils.RespondValidationError(c, "Invalid input")
		return
	}
	// Hash password before storing
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		utils.RespondInternalError(c, "Failed to hash password")
		return
	}
	user.Password = hashedPassword

	if err := config.DB.Create(&user).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": user})
}

// Login user
func Login(c *gin.Context) {
	var input models.User
	var user models.User

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondValidationError(c, "Invalid input")
		return
	}

	// Check if user exists
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		utils.RespondUnauthorized(c, "Invalid credentials")
		return
	}

	// Compare password
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		utils.RespondUnauthorized(c, "Invalid credentials")
		return
	}
	// Generate JWT Token
	userIDStr := fmt.Sprintf("%d", user.ID)
	token, err := utils.GenerateToken(userIDStr)
	if err != nil {
		utils.RespondInternalError(c, "Failed to generate token")
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user, "message": "Login successful"})
}

// Register user
func RegisterUserIfExistReturnUser(c *gin.Context) {
	var user models.User
	// Bind the JSON payload to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.RespondValidationError(c, "Invalid input")
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
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		utils.RespondInternalError(c, "Failed to hash password")
		return
	}
	user.Password = hashedPassword

	// Create the new user in the database
	if err := config.DB.Create(&user).Error; err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Failed to register user")
		return
	}

	// Return success response for newly registered user
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully", "user": user})
}
