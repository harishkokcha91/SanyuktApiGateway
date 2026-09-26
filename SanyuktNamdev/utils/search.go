package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SearchConfig holds search configuration for a model
type SearchConfig struct {
	SearchableFields []string
	SortableFields   []string
	DefaultSort      string
	MaxLimit         int
	DefaultLimit     int
}

// ProfileSearchConfig for Profile model
var ProfileSearchConfig = SearchConfig{
	SearchableFields: []string{"name", "current_location", "occupation", "address"},
	SortableFields:   []string{"created_at", "name", "status", "current_location"},
	DefaultSort:      "created_at",
	MaxLimit:         100,
	DefaultLimit:     20,
}

// BusinessSearchConfig for Business model
var BusinessSearchConfig = SearchConfig{
	SearchableFields: []string{"name", "description", "category", "location", "city", "state", "owner"},
	SortableFields:   []string{"created_at", "name", "status", "category", "location"},
	DefaultSort:      "created_at",
	MaxLimit:         100,
	DefaultLimit:     20,
}

// EventSearchConfig for Event model
var EventSearchConfig = SearchConfig{
	SearchableFields: []string{"name", "description", "category", "venue", "city", "state", "organizer"},
	SortableFields:   []string{"created_at", "name", "status", "event_date", "category", "city"},
	DefaultSort:      "created_at",
	MaxLimit:         100,
	DefaultLimit:     20,
}

// AchievementSearchConfig for Achievement model
var AchievementSearchConfig = SearchConfig{
	SearchableFields: []string{"name", "description", "achievement", "achievement_type"},
	SortableFields:   []string{"created_at", "name", "status", "date_of_achievement", "achievement_type"},
	DefaultSort:      "created_at",
	MaxLimit:         100,
	DefaultLimit:     20,
}

// SearchParams holds parsed search parameters
type SearchParams struct {
	Keyword     string
	Status      string
	SortBy      string
	SortOrder   string
	Limit       int
	Offset      int
	CustomFilters map[string]string
}

// ParseSearchParams extracts and validates search parameters from query
func ParseSearchParams(c *gin.Context, config SearchConfig, customFilters map[string]string) SearchParams {
	keyword := strings.TrimSpace(c.DefaultQuery("keyword", ""))
	status := strings.TrimSpace(c.DefaultQuery("status", ""))

	// Parse sort
	sortBy := strings.TrimSpace(c.DefaultQuery("sort_by", ""))
	sortOrder := strings.ToUpper(strings.TrimSpace(c.DefaultQuery("sort_order", "DESC")))

	// Validate sort_by against allowed fields
	if sortBy == "" {
		sortBy = config.DefaultSort
	} else {
		validSort := false
		for _, field := range config.SortableFields {
			if strings.EqualFold(field, sortBy) {
				sortBy = field
				validSort = true
				break
			}
		}
		if !validSort {
			sortBy = config.DefaultSort
		}
	}

	// Validate sort_order
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	// Parse pagination
	limit := config.DefaultLimit
	if l := c.DefaultQuery("limit", ""); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= config.MaxLimit {
			limit = parsed
		}
	}

	page := 1
	if p := c.DefaultQuery("page", ""); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	offset := (page - 1) * limit

	// Parse custom filters
	filters := make(map[string]string)
	for key, fieldName := range customFilters {
		if val := strings.TrimSpace(c.DefaultQuery(key, "")); val != "" {
			filters[fieldName] = val
		}
	}

	return SearchParams{
		Keyword:       keyword,
		Status:        status,
		SortBy:        sortBy,
		SortOrder:     sortOrder,
		Limit:         limit,
		Offset:        offset,
		CustomFilters: filters,
	}
}

// BuildSearchQuery builds a GORM query with search, filter, sort, and pagination
func BuildSearchQuery(db *gorm.DB, params SearchParams, config SearchConfig) *gorm.DB {
	query := db

	// Apply keyword search (ILIKE on searchable fields for PostgreSQL, LIKE for SQLite)
	if params.Keyword != "" {
		keyword := "%" + sanitizeForILIKE(params.Keyword) + "%"
		conditions := make([]string, 0, len(config.SearchableFields))
		args := make([]interface{}, 0, len(config.SearchableFields))

		for _, field := range config.SearchableFields {
			// Use LOWER() for case-insensitive search that works on both PostgreSQL and SQLite
			conditions = append(conditions, fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field))
			args = append(args, keyword)
		}

		query = query.Where(strings.Join(conditions, " OR "), args...)
	}

	// Apply status filter
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// Apply custom filters (exact match)
	for field, value := range params.CustomFilters {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	// Apply sorting
	orderClause := fmt.Sprintf("%s %s", params.SortBy, params.SortOrder)
	query = query.Order(orderClause)

	// Apply pagination
	query = query.Limit(params.Limit).Offset(params.Offset)

	return query
}

// BuildCountQuery builds a GORM query for counting total records (without pagination)
func BuildCountQuery(db *gorm.DB, params SearchParams, config SearchConfig) *gorm.DB {
	query := db

	// Apply keyword search
	if params.Keyword != "" {
		keyword := "%" + sanitizeForILIKE(params.Keyword) + "%"
		conditions := make([]string, 0, len(config.SearchableFields))
		args := make([]interface{}, 0, len(config.SearchableFields))

		for _, field := range config.SearchableFields {
			conditions = append(conditions, fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", field))
			args = append(args, keyword)
		}

		query = query.Where(strings.Join(conditions, " OR "), args...)
	}

	// Apply status filter
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// Apply custom filters
	for field, value := range params.CustomFilters {
		query = query.Where(fmt.Sprintf("%s = ?", field), value)
	}

	return query
}

// sanitizeForILIKE escapes special characters for ILIKE queries
func sanitizeForILIKE(input string) string {
	// Escape % and _ for literal search
	input = strings.ReplaceAll(input, "%", "\\%")
	input = strings.ReplaceAll(input, "_", "\\_")
	// Remove any potential SQL injection attempts
	input = strings.ReplaceAll(input, "'", "''")
	return input
}

// ProfileCustomFilters returns custom filter mappings for Profile
func ProfileCustomFilters() map[string]string {
	return map[string]string{
		"location":   "current_location",
		"age_range":  "age_range", // Special handling needed
	}
}

// BusinessCustomFilters returns custom filter mappings for Business
func BusinessCustomFilters() map[string]string {
	return map[string]string{
		"category": "category",
		"location": "location",
		"city":     "city",
		"state":    "state",
		"owner":    "owner",
	}
}

// EventCustomFilters returns custom filter mappings for Event
func EventCustomFilters() map[string]string {
	return map[string]string{
		"category": "category",
		"location": "city",
		"city":     "city",
		"state":    "state",
		"organizer": "organizer",
	}
}

// AchievementCustomFilters returns custom filter mappings for Achievement
func AchievementCustomFilters() map[string]string {
	return map[string]string{
		"type": "achievement_type",
	}
}

// ParseAgeRange parses age_range parameter (format: "min-max" or "min+" or "-max")
func ParseAgeRange(ageRange string) (minAge, maxAge *int) {
	if ageRange == "" {
		return nil, nil
	}

	parts := strings.Split(ageRange, "-")
	if len(parts) == 2 {
		if parts[0] != "" {
			if min, err := strconv.Atoi(parts[0]); err == nil && min > 0 {
				minAge = &min
			}
		}
		if parts[1] != "" {
			if max, err := strconv.Atoi(parts[1]); err == nil && max > 0 {
				maxAge = &max
			}
		}
	} else if len(parts) == 1 {
		if strings.HasSuffix(ageRange, "+") {
			minStr := strings.TrimSuffix(ageRange, "+")
			if min, err := strconv.Atoi(minStr); err == nil && min > 0 {
				minAge = &min
			}
		} else if strings.HasPrefix(ageRange, "-") {
			maxStr := strings.TrimPrefix(ageRange, "-")
			if max, err := strconv.Atoi(maxStr); err == nil && max > 0 {
				maxAge = &max
			}
		}
	}
	return minAge, maxAge
}

// ParseDateRange parses date_range parameter (format: "YYYY-MM-DD,YYYY-MM-DD")
func ParseDateRange(dateRange string) (startDate, endDate *time.Time) {
	if dateRange == "" {
		return nil, nil
	}

	parts := strings.Split(dateRange, ",")
	if len(parts) == 2 {
		if parts[0] != "" {
			if start, err := time.Parse("2006-01-02", parts[0]); err == nil {
				startDate = &start
			}
		}
		if parts[1] != "" {
			if end, err := time.Parse("2006-01-02", parts[1]); err == nil {
				endDate = &end
			}
		}
	}
	return startDate, endDate
}

// ApplyAgeRangeFilter applies age range filter to query (requires date_of_birth field)
func ApplyAgeRangeFilter(query *gorm.DB, minAge, maxAge *int) *gorm.DB {
	if minAge != nil {
		// Calculate max birth date (today - minAge years)
		maxBirthDate := time.Now().AddDate(-*minAge, 0, 0)
		query = query.Where("date_of_birth <= ?", maxBirthDate.Format("2006-01-02"))
	}
	if maxAge != nil {
		// Calculate min birth date (today - maxAge years)
		minBirthDate := time.Now().AddDate(-*maxAge, 0, 0)
		query = query.Where("date_of_birth >= ?", minBirthDate.Format("2006-01-02"))
	}
	return query
}

// ApplyDateRangeFilter applies date range filter to query
func ApplyDateRangeFilter(query *gorm.DB, fieldName string, startDate, endDate *time.Time) *gorm.DB {
	if startDate != nil {
		query = query.Where(fmt.Sprintf("%s >= ?", fieldName), startDate.Format("2006-01-02"))
	}
	if endDate != nil {
		query = query.Where(fmt.Sprintf("%s <= ?", fieldName), endDate.Format("2006-01-02"))
	}
	return query
}