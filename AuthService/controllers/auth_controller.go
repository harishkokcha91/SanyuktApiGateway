package controllers

import (
	"auth-service/config"
	"auth-service/models"
	"auth-service/utils"
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
	fmt.Println(user)
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
	fmt.Println(user)
	// Generate JWT Token
	userIDStr := fmt.Sprintf("%d", user.ID)
	token, _ := utils.GenerateToken(userIDStr)
	// claims, err := utils.ValidateToken(token)
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
	// 	c.Abort()
	// 	return
	// }

	c.JSON(http.StatusOK, gin.H{"token": token})
}
