package controllers

import (
	"math"
	"namdev-events/models"
	"net/http"
	"strconv"
	"time"

	initializers "namdev-events/database"

	"github.com/gin-gonic/gin"
)

// Get all events with pagination
func GetEvents(c *gin.Context) {
	var events []models.Event

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	if err := initializers.DB.Offset(offset).Limit(limit).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	var totalRecords int64
	initializers.DB.Model(&models.Event{}).Count(&totalRecords)

	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"page":         page,
		"limit":        limit,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
		"data":         events,
	})
}

// Get event by ID
func GetEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := initializers.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// Create a new event
func CreateEvent(c *gin.Context) {
	var event models.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	initializers.DB.Create(&event)
	c.JSON(http.StatusCreated, event)
}

// Update event
func UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := initializers.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.UpdatedAt = time.Now()
	initializers.DB.Save(&event)

	c.JSON(http.StatusOK, event)
}

// Delete event
func DeleteEvent(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := initializers.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	initializers.DB.Delete(&event)
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}
