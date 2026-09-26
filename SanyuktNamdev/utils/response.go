package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondError sends a JSON error response with the given status code
func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{Error: message})
}

// RespondDBError maps common GORM errors to appropriate HTTP status codes
func RespondDBError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	switch err {
	case gorm.ErrRecordNotFound:
		RespondError(c, http.StatusNotFound, "Resource not found")
	case gorm.ErrDuplicatedKey:
		RespondError(c, http.StatusConflict, "Resource already exists")
	default:
		RespondError(c, http.StatusInternalServerError, "Internal server error")
	}
}

// RespondValidationError sends a 400 validation error
func RespondValidationError(c *gin.Context, message string) {
	RespondError(c, http.StatusBadRequest, message)
}

// RespondUnauthorized sends a 401 unauthorized error
func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, message)
}

// RespondForbidden sends a 403 forbidden error
func RespondForbidden(c *gin.Context, message string) {
	RespondError(c, http.StatusForbidden, message)
}

// RespondNotFound sends a 404 not found error
func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, message)
}

// RespondInternalError sends a 500 internal server error
func RespondInternalError(c *gin.Context, message string) {
	RespondError(c, http.StatusInternalServerError, message)
}