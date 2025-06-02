package controllers

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
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

	// Fetch the profiles from the database with pagination
	if err := database.DB.Where("user_id = ?", id).Offset(offset).Limit(limitInt).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profiles"})
		return
	}

	// If no profiles are found, return a "not found" response
	if len(users) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get the total number of matching records
	var totalRecords int64
	if err := database.DB.Model(&models.Profile{}).Where("user_id = ?", id).Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch total records"})
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(limitInt)))

	// Return the paginated data as a response
	c.JSON(http.StatusOK, gin.H{
		"page":         pageInt,
		"limit":        limitInt,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
		"data":         users,
	})
}

func CreateUser(c *gin.Context) {
	var user models.Profile
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.Status = "pending"
	database.DB.Create(&user)
	c.JSON(http.StatusCreated, user)
}

func UploadImageForUser(c *gin.Context) {
	fmt.Println("UploadImageForUser")
	// Get user ID from URL params
	userID := c.Param("id")
	fmt.Println(userID)
	// Upload the image
	imagePath, err := UploadImage(c, userID)
	if err != nil {
		fmt.Println("err ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update user with image path
	if err := database.DB.Model(&models.Profile{}).Where("id = ?", userID).Update("image", imagePath).Error; err != nil {
		fmt.Println("Failed to update user image ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image uploaded successfully", "imagePath": imagePath})
}

// UploadImage handles image upload and returns the file path
func UploadImage(c *gin.Context, userID string) (string, error) {
	// Get file from form
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		return "", fmt.Errorf("image upload failed: %v", err)
	}
	defer file.Close()

	// Create uploads directory if not exists
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, os.ModePerm)
	}

	// Generate file name with user ID as suffix
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", os.Getpid(), userID, ext)
	filepath := filepath.Join(uploadDir, filename)

	// Save file
	outFile, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("could not save file: %v", err)
	}
	defer outFile.Close()

	// Copy file data to the new file
	if _, err = io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("failed to save image: %v", err)
	}

	return filepath, nil
}

func CreateUserWithImage(c *gin.Context) {
	// Parse form data
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	// Get user data
	var user models.Profile
	user.Name = c.PostForm("name")
	user.DateOfBirth = c.PostForm("dateOfBirth")
	user.BirthPlace = c.PostForm("birthPlace")

	// Upload image
	imagePath, err := UploadImage(c, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.Image = imagePath

	// Save user in database
	database.DB.Create(&user)

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
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
	var existingUser models.Profile

	// Fetch the existing user
	if err := database.DB.First(&existingUser, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Create a new struct to hold updates
	var updatedData models.Profile
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only non-empty fields
	database.DB.Model(&existingUser).Updates(updatedData)

	c.JSON(http.StatusOK, existingUser)
}

func DeleteUser(c *gin.Context) {
	fmt.Println("DeleteUser called")
	id := c.Param("id")
	if err := database.DB.Delete(&models.Profile{}, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
