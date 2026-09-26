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

// Get all events with pagination
func GetEvents(c *gin.Context) {
	var events []models.Event

	// Parse search parameters
	params := utils.ParseSearchParams(c, utils.EventSearchConfig, utils.EventCustomFilters())

	// Handle date_range special filter
	if dateRange := c.DefaultQuery("date_range", ""); dateRange != "" {
		params.CustomFilters["date_range"] = dateRange
	}

	// Build base query with public status filter
	baseQuery := utils.ApplyPublicStatusFilter(c, initializers.DB, "event")

	// Apply date range filter if provided
	if dateRange := c.DefaultQuery("date_range", ""); dateRange != "" {
		startDate, endDate := utils.ParseDateRange(dateRange)
		baseQuery = utils.ApplyDateRangeFilter(baseQuery, "event_date", startDate, endDate)
	}

	// Build search query with filters, sorting, pagination
	query := utils.BuildSearchQuery(baseQuery, params, utils.EventSearchConfig)

	if err := query.Find(&events).Error; err != nil {
		utils.RespondInternalError(c, "Failed to fetch events")
		return
	}

	// Get the total number of records (with filter applied)
	countQuery := utils.ApplyPublicStatusFilter(c, initializers.DB.Model(&models.Event{}), "event")
	if dateRange := c.DefaultQuery("date_range", ""); dateRange != "" {
		startDate, endDate := utils.ParseDateRange(dateRange)
		countQuery = utils.ApplyDateRangeFilter(countQuery, "event_date", startDate, endDate)
	}
	countQuery = utils.BuildCountQuery(countQuery, params, utils.EventSearchConfig)

	var totalRecords int64
	if err := countQuery.Count(&totalRecords).Error; err != nil {
		utils.RespondInternalError(c, "Failed to count events")
		return
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(params.Limit)))

	c.JSON(http.StatusOK, gin.H{
		"page":         params.Offset/params.Limit + 1,
		"limit":        params.Limit,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
		"data":         events,
	})
}

// Get event by ID
func GetEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	// First, try to find the event without status filter to check ownership
	if err := initializers.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	// Check if user is admin
	userRole, hasRole := c.Get("role")
	if hasRole && userRole == "admin" {
		c.JSON(http.StatusOK, event)
		return
	}

	// Public users can only see Approved events
	if event.Status != "Approved" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// Create a new event
func CreateEvent(c *gin.Context) {
	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Force Pending status - ignore any client-provided status
	event.Status = "Pending"
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	if err := initializers.DB.Create(&event).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusCreated, event)
}

// Update event
func UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := initializers.DB.First(&event, id).Error; err != nil {
		utils.RespondNotFound(c, "Event not found")
		return
	}

	// Get current user info from context
	userRole, hasUserRole := c.Get("role")

	// Ownership check: Event doesn't have a direct owner field
	// For now, only admins can update events
	// If we add an organizer/owner field later, we can add ownership check here
	isAdmin := hasUserRole && userRole == "admin"
	if !isAdmin {
		utils.RespondForbidden(c, "Admin access required to update event")
		return
	}

	var updatedData models.Event
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		utils.RespondValidationError(c, err.Error())
		return
	}

	// Owner edit on Approved record -> reset to Pending
	// Event uses "Approved" as approved status, "Pending" as pending status
	isApproved := event.Status == "Approved"

	// Prevent non-admins from directly setting status to Approved/Rejected
	if !isAdmin {
		if updatedData.Status == "Approved" || updatedData.Status == "Rejected" {
			updatedData.Status = event.Status
		}
	}

	event.UpdatedAt = time.Now()
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
		"name":                 updatedData.Name,
		"description":          updatedData.Description,
		"event_date":           updatedData.EventDate,
		"venue":                updatedData.Venue,
		"address":              updatedData.Address,
		"city":                 updatedData.City,
		"state":                updatedData.State,
		"zip_code":             updatedData.ZipCode,
		"country":              updatedData.Country,
		"organizer":            updatedData.Organizer,
		"email":                updatedData.Email,
		"phone":                updatedData.Phone,
		"category":             updatedData.Category,
		"capacity":             updatedData.Capacity,
		"attendees":            updatedData.Attendees,
		"status":               updatedData.Status,
		"image":                updatedData.Image,
		"reg_link":             updatedData.RegLink,
		"is_online":            updatedData.IsOnline,
		"ticket_price":         updatedData.TicketPrice,
	}

	// Only include audit fields in updates when owner resets to pending
	if shouldResetToPending {
		updates["approved_by"] = nil
		updates["approved_at"] = nil
		// RejectedBy/RejectedAt are preserved (not included in updates)
	}

	if err := initializers.DB.Model(&event).Updates(updates).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	// Re-fetch to return updated data
	if err := initializers.DB.First(&event, id).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}

	c.JSON(http.StatusOK, event)
}

// Delete event
func DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := initializers.DB.First(&event, id).Error; err != nil {
		utils.RespondNotFound(c, "Event not found")
		return
	}

	if err := initializers.DB.Delete(&event).Error; err != nil {
		utils.RespondDBError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}
