package notifications

import (
	"fmt"
	"net/http"
	"strconv"

	"SanyuktNamdev/database"
	"SanyuktNamdev/middleware"
	"SanyuktNamdev/models"
	"SanyuktNamdev/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Controller handles notification HTTP endpoints
type Controller struct {
	service *Service
}

// NewController creates a new notification controller
func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

// GetNotifications handles GET /notifications
// Query params: page, limit, unread_only
func (c *Controller) GetNotifications(ginCtx *gin.Context) {
	userID, exists := ginCtx.Get("userid")
	if !exists {
		utils.RespondUnauthorized(ginCtx, "User ID not found in token")
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	uid, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondValidationError(ginCtx, "Invalid user ID")
		return
	}

	page := 1
	if p := ginCtx.DefaultQuery("page", "1"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := ginCtx.DefaultQuery("limit", "20"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	unreadOnly := ginCtx.DefaultQuery("unread_only", "false") == "true"

	notifications, total, err := c.service.GetInAppNotifications(uint(uid), page, limit, unreadOnly)
	if err != nil {
		utils.RespondInternalError(ginCtx, "Failed to fetch notifications")
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	ginCtx.JSON(http.StatusOK, gin.H{
		"page":         page,
		"limit":        limit,
		"totalPages":   totalPages,
		"totalRecords": total,
		"data":         notifications,
	})
}

// GetUnreadCount handles GET /notifications/unread-count
func (c *Controller) GetUnreadCount(ginCtx *gin.Context) {
	userID, exists := ginCtx.Get("userid")
	if !exists {
		utils.RespondUnauthorized(ginCtx, "User ID not found in token")
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	uid, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondValidationError(ginCtx, "Invalid user ID")
		return
	}

	var count int64
	if err := database.DB.Model(&models.Notification{}).
		Where("user_id = ? AND channel = ? AND status = ?", uid, models.ChannelInApp, models.StatusPending).
		Count(&count).Error; err != nil {
		utils.RespondInternalError(ginCtx, "Failed to count unread notifications")
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// MarkAsRead handles PATCH /notifications/:id/read and PATCH /notifications/:id/mark-read
func (c *Controller) MarkAsRead(ginCtx *gin.Context) {
	userID, exists := ginCtx.Get("userid")
	if !exists {
		utils.RespondUnauthorized(ginCtx, "User ID not found in token")
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	uid, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondValidationError(ginCtx, "Invalid user ID")
		return
	}

	idStr := ginCtx.Param("id")
	nid, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.RespondValidationError(ginCtx, "Invalid notification ID")
		return
	}

	if err := c.service.MarkAsRead(uint(nid), uint(uid)); err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondNotFound(ginCtx, "Notification not found")
			return
		}
		utils.RespondInternalError(ginCtx, "Failed to mark notification as read")
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// SendTestNotification handles POST /notifications/test (admin only)
func (c *Controller) SendTestNotification(ginCtx *gin.Context) {
	// Verify admin role
	role, exists := ginCtx.Get("role")
	if !exists || role != "admin" {
		utils.RespondForbidden(ginCtx, "Admin access required")
		return
	}

	var req struct {
		UserID  uint   `json:"user_id" binding:"required"`
		Type    string `json:"type" binding:"required"`
		Message string `json:"message" binding:"required"`
		Channel string `json:"channel" binding:"required"`
	}

	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		utils.RespondValidationError(ginCtx, err.Error())
		return
	}

	// Validate channel
	validChannels := map[string]bool{
		string(models.ChannelInApp): true,
		string(models.ChannelEmail): true,
		string(models.ChannelPush):  true,
	}
	if !validChannels[req.Channel] {
		utils.RespondValidationError(ginCtx, "Invalid channel: must be 'in-app', 'email', or 'push'")
		return
	}

	notification, err := c.service.QueueNotification(req.UserID, req.Type, models.NotificationChannel(req.Channel), req.Message)
	if err != nil {
		utils.RespondInternalError(ginCtx, "Failed to create test notification")
		return
	}

	ginCtx.JSON(http.StatusCreated, gin.H{
		"message":      "Test notification sent",
		"notification": notification,
	})
}

// MarkAllAsRead handles PATCH /notifications/read-all
func (c *Controller) MarkAllAsRead(ginCtx *gin.Context) {
	userID, exists := ginCtx.Get("userid")
	if !exists {
		utils.RespondUnauthorized(ginCtx, "User ID not found in token")
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	uid, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		utils.RespondValidationError(ginCtx, "Invalid user ID")
		return
	}

	if err := c.service.MarkAllAsRead(uint(uid)); err != nil {
		utils.RespondInternalError(ginCtx, "Failed to mark all notifications as read")
		return
	}

	ginCtx.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// RegisterRoutes registers notification routes
func RegisterRoutes(r *gin.Engine, service *Service) {
	controller := NewController(service)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/notifications", controller.GetNotifications)
		protected.GET("/notifications/unread-count", controller.GetUnreadCount)
		protected.PATCH("/notifications/:id/read", controller.MarkAsRead)
		protected.PATCH("/notifications/:id/mark-read", controller.MarkAsRead)
		protected.PATCH("/notifications/read-all", controller.MarkAllAsRead)

		// Admin-only test endpoint
		adminOnly := protected.Group("/")
		adminOnly.Use(middleware.RequireRole("admin"))
		{
			adminOnly.POST("/notifications/test", controller.SendTestNotification)
		}
	}
}