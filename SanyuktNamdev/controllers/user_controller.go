package controllers

import (
	"SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetProfiles(c *gin.Context) {
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

func GetProfileByID(c *gin.Context) {
	id := c.Param("id")
	var user models.Profile
	if err := database.DB.First(&user, id).Error; err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	// Ownership check: only the profile owner can view
	userID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	userIDStr := fmt.Sprintf("%v", userID)
	if fmt.Sprintf("%d", user.UserId) != userIDStr {
		utils.RespondForbidden(c, "You can only view your own profile")
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetProfilesByUserID(c *gin.Context) {
	fmt.Println("GetProfilesByUserID")
	id := c.Param("id") // Get user ID from the URL
	var users []models.Profile

	// Ownership check: only the profile owner can view their profiles
	userID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	userIDStr := fmt.Sprintf("%v", userID)
	if id != userIDStr {
		utils.RespondForbidden(c, "You can only view your own profiles")
		return
	}

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

	// Calculate the offset based on the page number
	offset := (pageInt - 1) * limitInt

	// Fetch the profiles from the database with pagination
	if err := database.DB.Where("user_id = ?", id).Offset(offset).Limit(limitInt).Find(&users).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch profiles")
		return
	}

	// If no profiles are found, return a "not found" response
	if len(users) == 0 {
		utils.RespondNotFound(c, "User not found")
		return
	}

	// Get the total number of matching records
	var totalRecords int64
	if err := database.DB.Model(&models.Profile{}).Where("user_id = ?", id).Count(&totalRecords).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch total records")
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

func CreateProfile(c *gin.Context) {
	var user models.Profile
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Set UserId from authenticated user
	authUserID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	user.UserId = authUserID.(uint)

	user.Status = "pending"
	if err := database.DB.Create(&user).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

func UploadProfileImage(c *gin.Context) {
	fmt.Println("UploadProfileImage")
	// Get user ID from URL params
	userID := c.Param("id")
	fmt.Println(userID)

	// Ownership check: only the profile owner can upload image
	authUserID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	authUserIDStr := fmt.Sprintf("%v", authUserID)
	if userID != authUserIDStr {
		utils.RespondForbidden(c, "You can only upload image for your own profile")
		return
	}

	// Upload the image
	imagePath, err := UploadImageForProfile(c, userID)
	if err != nil {
		fmt.Println("err ", err)
		utils.RespondInternalError(c, err.Error())
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

// UploadImageForProfile handles image upload and returns the file path
func UploadImageForProfile(c *gin.Context, userID string) (string, error) {
	// Get file from form
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		return "", fmt.Errorf("image upload failed: %v", err)
	}
	defer file.Close()

	// Read first 512 bytes for MIME detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for MIME detection: %v", err)
	}
	// Reset reader
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file reader: %v", err)
	}

	// MIME type validation
	mimeType := http.DetectContentType(buffer[:n])
	allowedMimes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
	}
	if !allowedMimes[mimeType] {
		return "", fmt.Errorf("invalid file type: %s. Only JPEG and PNG allowed", mimeType)
	}

	// Extension from MIME (not from user-provided filename)
	var ext string
	switch mimeType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	}

	// Create uploads directory if not exists
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create upload directory: %v", err)
		}
	}

	// UUID-based filename (unpredictable, no collisions)
	uuidStr := uuid.New().String()
	filename := fmt.Sprintf("%s%s", uuidStr, ext)
	filePath := filepath.Join(uploadDir, filename)

	// Extra safety: ensure path stays within uploadDir
	absUploadDir, _ := filepath.Abs(uploadDir)
	absFilePath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFilePath, absUploadDir) {
		return "", fmt.Errorf("invalid file path")
	}

	// Save file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("could not save file: %v", err)
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("failed to save image: %v", err)
	}

	return filePath, nil
}

func CreateProfileWithImage(c *gin.Context) {
	// Parse form data
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		utils.RespondValidationError(c, "File too large")
		return
	}

	// Get user data
	var user models.Profile
	user.Name = c.PostForm("name")
	user.DateOfBirth = c.PostForm("dateOfBirth")
	user.BirthPlace = c.PostForm("birthPlace")

	// Set UserId from authenticated user
	authUserID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	user.UserId = authUserID.(uint)

	// Upload image
	imagePath, err := UploadImageForProfile(c, fmt.Sprintf("%d", user.UserId))
	if err != nil {
		utils.RespondInternalError(c, err.Error())
		return
	}
	user.Image = imagePath

	// Save user in database
	if err := database.DB.Create(&user).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
}

func convertProfileDateFormat(inputDate string) (time.Time, error) {
	const customDateFormat = "2006-01-02" // Format for date without time component

	// Parse the date from JSON
	dateOfBirth, err := time.Parse(customDateFormat, inputDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return dateOfBirth, nil
}
func UpdateProfile(c *gin.Context) {
	id := c.Param("id")
	var existingUser models.Profile

	// Fetch the existing user
	if err := database.DB.First(&existingUser, id).Error; err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	// Ownership check: only the profile owner can update
	userID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	userIDStr := fmt.Sprintf("%v", userID)
	if fmt.Sprintf("%d", existingUser.UserId) != userIDStr {
		utils.RespondForbidden(c, "You can only update your own profile")
		return
	}

	// Create a new struct to hold updates
	var updatedData models.Profile
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Update only non-empty fields
	if err := database.DB.Model(&existingUser).Updates(updatedData).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, existingUser)
}

func DeleteProfile(c *gin.Context) {
	fmt.Println("DeleteProfile called")
	id := c.Param("id")
	var existingUser models.Profile

	// Fetch the existing user
	if err := database.DB.First(&existingUser, id).Error; err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	// Ownership check: only the profile owner can delete
	userID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	userIDStr := fmt.Sprintf("%v", userID)
	if fmt.Sprintf("%d", existingUser.UserId) != userIDStr {
		utils.RespondForbidden(c, "You can only delete your own profile")
		return
	}

	if err := database.DB.Delete(&models.Profile{}, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ============================================
// User CRUD (merged from userController.go)
// ============================================

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
		utils.RespondInternalError(c, "Failed to fetch users")
		return
	}

	// Get the total number of records
	var totalRecords int64
	if err := database.DB.Model(&models.User{}).Count(&totalRecords).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch total records")
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

	// Ownership check: only the user themselves can view
	userID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	userIDStr := fmt.Sprintf("%v", userID)
	if fmt.Sprintf("%d", user.ID) != userIDStr {
		utils.RespondForbidden(c, "You can only view your own user record")
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

	// Ownership check: only the user themselves can update
	userID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	userIDStr := fmt.Sprintf("%v", userID)
	if fmt.Sprintf("%d", user.ID) != userIDStr {
		utils.RespondForbidden(c, "You can only update your own user record")
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
	var user models.User

	// Fetch user first for ownership check
	if err := database.DB.First(&user, id).Error; err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	// Ownership check: only the user themselves can delete
	userID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	userIDStr := fmt.Sprintf("%v", userID)
	if fmt.Sprintf("%d", user.ID) != userIDStr {
		utils.RespondForbidden(c, "You can only delete your own user record")
		return
	}

	if err := database.DB.Delete(&models.User{}, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
