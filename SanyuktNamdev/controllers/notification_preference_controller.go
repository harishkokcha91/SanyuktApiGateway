package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NotificationPreferenceController handles notification preference endpoints
type NotificationPreferenceController struct {
	db *gorm.DB
}

// NewNotificationPreferenceController creates a new notification preference controller
func NewNotificationPreferenceController() *NotificationPreferenceController {
	return &NotificationPreferenceController{
		db: database.DB,
	}
}

// GetNotificationPreferences handles GET /notification-preferences
func (c *NotificationPreferenceController) GetNotificationPreferences(ctx *gin.Context) {
	userID, exists := ctx.Get("userid")
	if !exists {
		utils.RespondUnauthorized(ctx, "User ID not found in token")
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	uid, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondInternalError(ctx, "Invalid user ID")
		return
	}

	var preferences []models.NotificationPreference
	if err := c.db.Where("user_id = ?", uid).Find(&preferences).Error; err != nil {
		utils.RespondDBError(ctx, err)
		return
	}

	// Ensure all channels have a preference (create defaults if missing)
	channels := []models.NotificationChannel{
		models.ChannelInApp,
		models.ChannelEmail,
		models.ChannelPush,
	}

	prefMap := make(map[models.NotificationChannel]*models.NotificationPreference)
	for i := range preferences {
		prefMap[preferences[i].Channel] = &preferences[i]
	}

	for _, channel := range channels {
		if _, exists := prefMap[channel]; !exists {
			// Create default preference
			newPref := models.NotificationPreference{
				UserID:  uint(uid),
				Channel: channel,
				Enabled: true,
			}
			if err := c.db.Create(&newPref).Error; err != nil {
				utils.RespondDBError(ctx, err)
				return
			}
			preferences = append(preferences, newPref)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Notification preferences retrieved successfully",
		"data":    preferences,
	})
}

// UpdateNotificationPreference handles PATCH /notification-preferences/:channel
func (c *NotificationPreferenceController) UpdateNotificationPreference(ctx *gin.Context) {
	userID, exists := ctx.Get("userid")
	if !exists {
		utils.RespondUnauthorized(ctx, "User ID not found in token")
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	uid, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondInternalError(ctx, "Invalid user ID")
		return
	}

	channel := ctx.Param("channel")
	validChannels := map[string]bool{
		"in-app": true,
		"email":  true,
		"push":   true,
	}

	if !validChannels[channel] {
		utils.RespondValidationError(ctx, "Invalid channel: must be one of in-app, email, push")
		return
	}

	var req struct {
		Enabled *bool `json:"enabled"`
	}

	if err = ctx.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(ctx, "Invalid request body: enabled (boolean) required")
		return
	}

	if req.Enabled == nil {
		utils.RespondValidationError(ctx, "Invalid request body: enabled (boolean) required")
		return
	}

	enabled := *req.Enabled

	var preference models.NotificationPreference
	err = c.db.Where("user_id = ? AND channel = ?", uid, channel).First(&preference).Error

	if err == gorm.ErrRecordNotFound {
		// Create new preference using map to avoid GORM zero-value skipping
		preferenceMap := map[string]interface{}{
			"user_id": uint(uid),
			"channel": channel,
			"enabled": enabled,
		}
		if err := c.db.Model(&models.NotificationPreference{}).Create(preferenceMap).Error; err != nil {
			utils.RespondDBError(ctx, err)
			return
		}
		// Fetch the created record to return it
		if err := c.db.Where("user_id = ? AND channel = ?", uid, channel).First(&preference).Error; err != nil {
			utils.RespondDBError(ctx, err)
			return
		}
	} else if err != nil {
		utils.RespondDBError(ctx, err)
		return
	} else {
		// Update existing
		preference.Enabled = enabled
		if err := c.db.Save(&preference).Error; err != nil {
			utils.RespondDBError(ctx, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Notification preference updated successfully",
		"data":    preference,
	})
}

// ResetNotificationPreferences handles POST /notification-preferences/reset
func (c *NotificationPreferenceController) ResetNotificationPreferences(ctx *gin.Context) {
	userID, exists := ctx.Get("userid")
	if !exists {
		utils.RespondUnauthorized(ctx, "User ID not found in token")
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	uid, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondInternalError(ctx, "Invalid user ID")
		return
	}

	channels := []models.NotificationChannel{
		models.ChannelInApp,
		models.ChannelEmail,
		models.ChannelPush,
	}

	for _, channel := range channels {
		var preference models.NotificationPreference
		err := c.db.Where("user_id = ? AND channel = ?", uid, channel).First(&preference).Error

		if err == gorm.ErrRecordNotFound {
			// Create new with default
			preference = models.NotificationPreference{
				UserID:  uint(uid),
				Channel: channel,
				Enabled: true,
			}
			if err := c.db.Create(&preference).Error; err != nil {
				utils.RespondDBError(ctx, err)
				return
			}
		} else if err != nil {
			utils.RespondDBError(ctx, err)
			return
		} else {
			// Reset to default
			preference.Enabled = true
			if err := c.db.Save(&preference).Error; err != nil {
				utils.RespondDBError(ctx, err)
				return
			}
		}
	}

	// Fetch updated preferences
	var preferences []models.NotificationPreference
	if err := c.db.Where("user_id = ?", uid).Find(&preferences).Error; err != nil {
		utils.RespondDBError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Notification preferences reset to defaults successfully",
		"data":    preferences,
	})
}