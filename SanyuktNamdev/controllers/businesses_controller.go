package controllers

import (
	initializers "SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Get all Businesses with pagination
func GetBusinesses(c *gin.Context) {
	var businesses []models.Business

	// Parse search parameters
	params := utils.ParseSearchParams(c, utils.BusinessSearchConfig, utils.BusinessCustomFilters())

	// Build base query with public status filter
	baseQuery := utils.ApplyPublicStatusFilter(c, initializers.DB, "business")

	// Build search query with filters, sorting, pagination
	query := utils.BuildSearchQuery(baseQuery, params, utils.BusinessSearchConfig)

	if err := query.Find(&businesses).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch businesses")
		return
	}

	// Get the total number of records (with filter applied)
	countQuery := utils.ApplyPublicStatusFilter(c, initializers.DB.Model(&models.Business{}), "business")
	countQuery = utils.BuildCountQuery(countQuery, params, utils.BusinessSearchConfig)

	var totalRecords int64
	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.RespondInternalError(c, "Failed to count businesses")
		return
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(params.Limit)))

	c.JSON(http.StatusOK, gin.H{
		"page":         params.Offset/params.Limit + 1,
		"limit":        params.Limit,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
		"data":         businesses,
	})
}

// Get business by ID
func GetBusinessByID(c *gin.Context) {
	id := c.Param("id")
	var business models.Business

	// First, try to find the business without status filter to check ownership
	if err := initializers.DB.First(&business, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	// Check if user is admin
	userRole, hasRole := c.Get("role")
	if hasRole && userRole == "admin" {
		c.JSON(http.StatusOK, business)
		return
	}

	// For business, there's no direct owner field like profiles
	// Public users can only see Approved businesses
	if business.Status != "Approved" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	c.JSON(http.StatusOK, business)
}

// Create a new business
func CreateBusiness(c *gin.Context) {
	var business models.Business

	if err := c.ShouldBindJSON(&business); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Force Pending status - ignore any client-provided status
	business.Status = "Pending"
	business.CreatedAt = time.Now()
	business.UpdatedAt = time.Now()

	if err := initializers.DB.Create(&business).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusCreated, business)
}

// Update business
func UpdateBusiness(c *gin.Context) {
	id := c.Param("id")
	var business models.Business

	if err := initializers.DB.First(&business, id).Error; err != nil {
		utils.RespondNotFound(c, "Business not found")
		return
	}

	// Get current user info from context
	userRole, hasUserRole := c.Get("role")

	// Ownership check: Business doesn't have a direct owner field
	// For now, only admins can update businesses
	// If we add an owner field later, we can add ownership check here
	isAdmin := hasUserRole && userRole == "admin"
	if !isAdmin {
		// For non-admins, we could check if they are the "owner" via a field
		// Currently, only admins can update
		utils.RespondForbidden(c, "Admin access required to update business")
		return
	}

	var updatedData models.Business
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Owner edit on Approved record -> reset to Pending
	// Business uses "Approved" as approved status, "Pending" as pending status
	isApproved := business.Status == "Approved"
	// Since only admins can update, this logic applies if we add ownership later
	// For now, admin edits don't reset status

	// Prevent non-admins from directly setting status to Approved/Rejected
	if !isAdmin {
		if updatedData.Status == "Approved" || updatedData.Status == "Rejected" {
			updatedData.Status = business.Status
		}
	}

	business.UpdatedAt = time.Now()
	// Preserve audit fields if status is being reset by owner
	if !isAdmin && isApproved {
		updatedData.Status = "Pending"
		updatedData.ApprovedBy = nil
		updatedData.ApprovedAt = nil
		// Preserve RejectedBy/RejectedAt if previously rejected
	}

	// Build update map to handle pointer fields correctly
	// Only include audit fields when owner resets to pending
	shouldResetToPending := !isAdmin && isApproved
	updates := map[string]interface{}{
		"name":             updatedData.Name,
		"category":         updatedData.Category,
		"description":      updatedData.Description,
		"owner":            updatedData.Owner,
		"email":            updatedData.Email,
		"phone":            updatedData.Phone,
		"whats_app":        updatedData.WhatsApp,
		"location":         updatedData.Location,
		"address":          updatedData.Address,
		"city":             updatedData.City,
		"state":            updatedData.State,
		"zip_code":         updatedData.ZipCode,
		"country":          updatedData.Country,
		"website":          updatedData.Website,
		"image":            updatedData.Image,
		"status":           updatedData.Status,
		"is_verified":      updatedData.IsVerified,
		"opening_hours":    updatedData.OpeningHours,
		"home_delivery":    updatedData.HomeDelivery,
		"payment_methods":  updatedData.PaymentMethods,
	}

	// Only include audit fields in updates when owner resets to pending
	if shouldResetToPending {
		updates["approved_by"] = nil
		updates["approved_at"] = nil
		// RejectedBy/RejectedAt are preserved (not included in updates)
	}

	if err := initializers.DB.Model(&business).Updates(updates).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	// Re-fetch to return updated data
	if err := initializers.DB.First(&business, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, business)
}

// Delete business
func DeleteBusiness(c *gin.Context) {
	id := c.Param("id")
	var business models.Business

	if err := initializers.DB.First(&business, id).Error; err != nil {
		utils.RespondNotFound(c, "Business not found")
		return
	}

	if err := initializers.DB.Delete(&business).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Business deleted successfully"})

	if err := initializers.DB.First(&business, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Business not found"})
		return
	}

	initializers.DB.Delete(&business)
	c.JSON(http.StatusOK, gin.H{"message": "Business deleted successfully"})
}
