package controllers

import (
	"SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/notifications"
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

	// Parse search parameters
	params := utils.ParseSearchParams(c, utils.ProfileSearchConfig, utils.ProfileCustomFilters())

	// Build base query with public status filter
	baseQuery := utils.ApplyPublicStatusFilter(c, database.DB, "profile")

	// Apply age range filter if provided
	if ageRange := c.DefaultQuery("age_range", ""); ageRange != "" {
		minAge, maxAge := utils.ParseAgeRange(ageRange)
		baseQuery = utils.ApplyAgeRangeFilter(baseQuery, minAge, maxAge)
	}

	// Build search query with filters, sorting, pagination
	query := utils.BuildSearchQuery(baseQuery, params, utils.ProfileSearchConfig)

	// Fetch the data from the database with pagination
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Get the total number of records (with filter applied)
	countQuery := utils.ApplyPublicStatusFilter(c, database.DB.Model(&models.Profile{}), "profile")
	if ageRange := c.DefaultQuery("age_range", ""); ageRange != "" {
		minAge, maxAge := utils.ParseAgeRange(ageRange)
		countQuery = utils.ApplyAgeRangeFilter(countQuery, minAge, maxAge)
	}
	countQuery = utils.BuildCountQuery(countQuery, params, utils.ProfileSearchConfig)

	var totalRecords int64
	if err := countQuery.Count(&totalRecords).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch total records"})
		return
	}

	// Calculate totalPages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(params.Limit)))

	// Return the paginated data as response
	c.JSON(http.StatusOK, gin.H{
		"page":         params.Offset/params.Limit + 1,
		"limit":        params.Limit,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
		"data":         users,
	})
}

func GetProfileByID(c *gin.Context) {
	id := c.Param("id")
	var user models.Profile

	// First, try to find the profile without status filter to check ownership
	if err := database.DB.First(&user, id).Error; err != nil {
		utils.RespondNotFound(c, "User not found")
		return
	}

	// Check if user is admin
	userRole, hasRole := c.Get("role")
	if hasRole && userRole == "admin" {
		c.JSON(http.StatusOK, user)
		return
	}

	// Check if user is the owner
	userID, exists := c.Get("userid")
	if !exists {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}
	userIDStr := fmt.Sprintf("%v", userID)
	if fmt.Sprintf("%d", user.UserId) != userIDStr {
		// Not owner - check if profile is public (active)
		if user.Status != "active" {
			utils.RespondNotFound(c, "User not found")
			return
		}
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

	// Parse search parameters
	params := utils.ParseSearchParams(c, utils.ProfileSearchConfig, utils.ProfileCustomFilters())

	// Build base query with user_id filter
	baseQuery := database.DB.Where("user_id = ?", id)

	// Apply age range filter if provided
	if ageRange := c.DefaultQuery("age_range", ""); ageRange != "" {
		minAge, maxAge := utils.ParseAgeRange(ageRange)
		baseQuery = utils.ApplyAgeRangeFilter(baseQuery, minAge, maxAge)
	}

	// Build search query with filters, sorting, pagination
	query := utils.BuildSearchQuery(baseQuery, params, utils.ProfileSearchConfig)

	// Fetch the profiles from the database with pagination
	if err := query.Find(&users).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch profiles")
		return
	}

	// If no profiles are found, return empty array (not "not found")
	// if len(users) == 0 {
	// 	utils.RespondNotFound(c, "User not found")
	// 	return
	// }

	// Get the total number of matching records
	countQuery := database.DB.Model(&models.Profile{}).Where("user_id = ?", id)
	if ageRange := c.DefaultQuery("age_range", ""); ageRange != "" {
		minAge, maxAge := utils.ParseAgeRange(ageRange)
		countQuery = utils.ApplyAgeRangeFilter(countQuery, minAge, maxAge)
	}
	countQuery = utils.BuildCountQuery(countQuery, params, utils.ProfileSearchConfig)

	var totalRecords int64
	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch total records")
		return
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(params.Limit)))

	// Return the paginated data as a response
	c.JSON(http.StatusOK, gin.H{
		"page":         params.Offset/params.Limit + 1,
		"limit":        params.Limit,
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
	userIDStr := fmt.Sprintf("%v", authUserID)
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondInternalError(c, "Invalid user ID")
		return
	}
	user.UserId = uint(userID)

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
	userIDStr := fmt.Sprintf("%v", authUserID)
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondInternalError(c, "Invalid user ID")
		return
	}
	user.UserId = uint(userID)

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
// UpdateProfile handles PUT /matrimonialProfiles/:id
func UpdateProfile(c *gin.Context) {
	id := c.Param("id")
	var existingProfile models.Profile

	// Fetch the existing profile
	if err := database.DB.First(&existingProfile, id).Error; err != nil {
		utils.RespondNotFound(c, "Profile not found")
		return
	}

	// Get current user info from context
	userID, hasUserID := c.Get("userid")
	userRole, hasUserRole := c.Get("role")

	// Ownership check: only the profile owner can update (admins can also update)
	if hasUserID {
		userIDStr := fmt.Sprintf("%v", userID)
		if fmt.Sprintf("%d", existingProfile.UserId) != userIDStr {
			// Not owner, check if admin
			if !(hasUserRole && userRole == "admin") {
				utils.RespondForbidden(c, "You can only update your own profile")
				return
			}
		}
	} else {
		utils.RespondUnauthorized(c, "User ID not found in token")
		return
	}

	// Create a new struct to hold updates
	var updatedData models.Profile
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Determine if current user is admin
	isAdmin := hasUserRole && userRole == "admin"

	// Owner edit on Approved/active record -> reset to Pending
	// Profile uses "active" as approved status, "pending" as pending status
	isApproved := existingProfile.Status == "active"
	shouldResetToPending := !isAdmin && hasUserID && isApproved
	if shouldResetToPending {
		// Owner editing an approved profile - reset to pending
		updatedData.Status = "pending"
		updatedData.ApprovedBy = nil
		updatedData.ApprovedAt = nil
		// Preserve RejectedBy/RejectedAt if previously rejected
	}

	// Prevent owners from directly setting status to Approved/Rejected
	// Only admins can set status to active (approved) or rejected
	if !isAdmin {
		if updatedData.Status == "active" || updatedData.Status == "inactive" {
			// Ignore owner-provided status changes to approved/rejected states
			updatedData.Status = existingProfile.Status
		}
	}

	// Build update map to handle pointer fields correctly (GORM doesn't update zero values in structs)
	// Only include audit fields when they're being intentionally changed (owner re-pending)
	updates := map[string]interface{}{
		"profile_for":        updatedData.ProfileFor,
		"name":               updatedData.Name,
		"image":              updatedData.Image,
		"date_of_birth":      updatedData.DateOfBirth,
		"birth_place":        updatedData.BirthPlace,
		"height":             updatedData.Height,
		"complexion":         updatedData.Complexion,
		"gotra_self":         updatedData.GotraSelf,
		"gotra_mother":       updatedData.GotraMother,
		"gotra_grand_mother": updatedData.GotraGrandMother,
		"manglik":            updatedData.Manglik,
		"father_name":        updatedData.FatherName,
		"father_occupation":  updatedData.FatherOccupation,
		"mother_name":        updatedData.MotherName,
		"mother_occupation":  updatedData.MotherOccupation,
		"siblings":           updatedData.Siblings,
		"qualification":      updatedData.Qualification,
		"occupation":         updatedData.Occupation,
		"annual_income":      updatedData.AnnualIncome,
		"marital_status":     updatedData.MaritalStatus,
		"address":            updatedData.Address,
		"current_location":   updatedData.CurrentLocation,
		"status":             updatedData.Status,
		"phone_numbers":      updatedData.PhoneNumbers,
	}

	// Only include audit fields in updates when owner resets to pending
	if shouldResetToPending {
		updates["approved_by"] = nil
		updates["approved_at"] = nil
		// RejectedBy/RejectedAt are preserved (not included in updates)
	}

	if err := database.DB.Model(&existingProfile).Updates(updates).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	// Notify admins if owner edited an approved profile (reset to pending)
	if shouldResetToPending {
		notifService := notifications.NewService()
		userIDStr := fmt.Sprintf("%v", userID)
		userIDVal, _ := strconv.ParseUint(userIDStr, 10, 64)
		var user models.User
		ownerName := "Unknown"
		if err := database.DB.First(&user, uint(userIDVal)).Error; err == nil {
			ownerName = user.Name
		}
		if err := notifService.NotifyOnOwnerEdit("Profile", existingProfile.Name, ownerName); err != nil {
			fmt.Printf("Failed to send owner edit notification: %v\n", err)
		}
	}

	// Re-fetch to return updated data
	if err := database.DB.First(&existingProfile, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, existingProfile)
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
