package utils

import (
	"SanyuktNamdev/database"
	"SanyuktNamdev/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// StatusCount represents count of items by status
type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// PendingItem represents an item pending approval
type PendingItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"`
}

// TurnaroundStats represents approval turnaround statistics
type TurnaroundStats struct {
	ModelType          string  `json:"model_type"`
	AverageTurnaround  float64 `json:"average_turnaround_seconds"`
	MinTurnaround      float64 `json:"min_turnaround_seconds"`
	MaxTurnaround      float64 `json:"max_turnaround_seconds"`
	TotalApproved      int64   `json:"total_approved"`
}

// RejectionRate represents rejection rate statistics
type RejectionRate struct {
	ModelType     string             `json:"model_type"`
	TotalRejected int64              `json:"total_rejected"`
	TotalCreated  int64              `json:"total_created"`
	RejectionRate float64            `json:"rejection_rate"`
	ByReason      []RejectionByReason `json:"by_reason"`
}

// RejectionByReason represents rejection counts grouped by reason
type RejectionByReason struct {
	Reason string `json:"reason"`
	Count  int64  `json:"count"`
}

// AnalyticsSummary represents overall analytics summary
type AnalyticsSummary struct {
	Profiles     []StatusCount `json:"profiles"`
	Businesses   []StatusCount `json:"businesses"`
	Events       []StatusCount `json:"events"`
	Achievements []StatusCount `json:"achievements"`
	TotalItems   int64         `json:"total_items"`
	TotalPending int64         `json:"total_pending"`
	TotalApproved int64        `json:"total_approved"`
	TotalRejected int64        `json:"total_rejected"`
}

// getDB returns the database instance
func getDB() *gorm.DB {
	return database.DB
}

// getModelTableName returns the table name for a given model type
func getModelTableName(modelType string) string {
	switch modelType {
	case "profile":
		return "profiles"
	case "business":
		return "businesses"
	case "event":
		return "events"
	case "achievement":
		return "achievements"
	default:
		return ""
	}
}

// getModelStatusValues returns the possible status values for a model type
func getModelStatusValues(modelType string) []string {
	switch modelType {
	case "profile":
		return []string{"pending", "active", "inactive"}
	case "business":
		return []string{"Pending", "Approved", "Rejected"}
	case "event":
		return []string{"Pending", "Approved", "Upcoming", "Completed", "Cancelled"}
	case "achievement":
		return []string{"Pending", "Approved", "Rejected"}
	default:
		return []string{}
	}
}

// GetStatusCounts returns counts grouped by status for a specific model type
func GetStatusCounts(modelType string) ([]StatusCount, error) {
	db := getDB()
	tableName := getModelTableName(modelType)
	if tableName == "" {
		return nil, fmt.Errorf("invalid model type: %s", modelType)
	}

	var results []StatusCount
	query := fmt.Sprintf("SELECT status, COUNT(*) as count FROM %s GROUP BY status", tableName)
	err := db.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Ensure all status values are represented (even with 0 count)
	statusValues := getModelStatusValues(modelType)
	found := make(map[string]bool)
	for _, r := range results {
		found[r.Status] = true
	}

	for _, status := range statusValues {
		if !found[status] {
			results = append(results, StatusCount{Status: status, Count: 0})
		}
	}

	return results, nil
}

// GetAllStatusCounts returns status counts for all model types
func GetAllStatusCounts() (AnalyticsSummary, error) {
	summary := AnalyticsSummary{}

	profiles, err := GetStatusCounts("profile")
	if err != nil {
		return summary, err
	}
	summary.Profiles = profiles

	businesses, err := GetStatusCounts("business")
	if err != nil {
		return summary, err
	}
	summary.Businesses = businesses

	events, err := GetStatusCounts("event")
	if err != nil {
		return summary, err
	}
	summary.Events = events

	achievements, err := GetStatusCounts("achievement")
	if err != nil {
		return summary, err
	}
	summary.Achievements = achievements

	// Calculate totals
	for _, p := range profiles {
		summary.TotalItems += p.Count
		if p.Status == "pending" {
			summary.TotalPending += p.Count
		} else if p.Status == "active" {
			summary.TotalApproved += p.Count
		} else if p.Status == "inactive" {
			summary.TotalRejected += p.Count
		}
	}
	for _, b := range businesses {
		summary.TotalItems += b.Count
		if b.Status == "Pending" {
			summary.TotalPending += b.Count
		} else if b.Status == "Approved" {
			summary.TotalApproved += b.Count
		} else if b.Status == "Rejected" {
			summary.TotalRejected += b.Count
		}
	}
	for _, e := range events {
		summary.TotalItems += e.Count
		if e.Status == "Pending" {
			summary.TotalPending += e.Count
		} else if e.Status == "Approved" {
			summary.TotalApproved += e.Count
		} else if e.Status == "Rejected" {
			summary.TotalRejected += e.Count
		}
	}
	for _, a := range achievements {
		summary.TotalItems += a.Count
		if a.Status == "Pending" {
			summary.TotalPending += a.Count
		} else if a.Status == "Approved" {
			summary.TotalApproved += a.Count
		} else if a.Status == "Rejected" {
			summary.TotalRejected += a.Count
		}
	}

	return summary, nil
}

// GetPendingItems returns paginated list of items pending approval
func GetPendingItems(modelType string, limit, offset int) ([]PendingItem, int64, error) {
	db := getDB()
	tableName := getModelTableName(modelType)
	if tableName == "" {
		return nil, 0, fmt.Errorf("invalid model type: %s", modelType)
	}

	var pendingStatus string
	switch modelType {
	case "profile":
		pendingStatus = "pending"
	case "business", "event", "achievement":
		pendingStatus = "Pending"
	default:
		return nil, 0, fmt.Errorf("invalid model type: %s", modelType)
	}

	var total int64
	if err := db.Table(tableName).Where("status = ?", pendingStatus).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []PendingItem
	// Use GORM for better database compatibility
	err := db.Table(tableName).
		Select("id, name, status, created_at").
		Where("status = ?", pendingStatus).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}

	// Add type to each item
	for i := range items {
		items[i].Type = modelType
	}

	return items, total, nil
}

// GetApprovalTurnaround returns average approval turnaround time for a model type
func GetApprovalTurnaround(modelType string) (TurnaroundStats, error) {
	db := getDB()
	tableName := getModelTableName(modelType)
	if tableName == "" {
		return TurnaroundStats{}, fmt.Errorf("invalid model type: %s", modelType)
	}

	// Only calculate for items that have been approved
	var stats TurnaroundStats
	stats.ModelType = modelType

	// Use database-agnostic approach for turnaround calculation
	// Use a simpler approach: fetch data and calculate in Go for database compatibility
	var rows []struct {
		CreatedAt  time.Time
		ApprovedAt time.Time
	}

	err := db.Table(tableName).
		Select("created_at, approved_at").
		Where("approved_at IS NOT NULL").
		Where("status IN (?, ?)", "Approved", "active").
		Scan(&rows).Error
	if err != nil {
		return stats, err
	}

	stats.TotalApproved = int64(len(rows))
	if stats.TotalApproved > 0 {
		var sum, min, max float64
		for i, row := range rows {
			diff := row.ApprovedAt.Sub(row.CreatedAt).Seconds()
			sum += diff
			if i == 0 || diff < min {
				min = diff
			}
			if i == 0 || diff > max {
				max = diff
			}
		}
		stats.AverageTurnaround = sum / float64(len(rows))
		stats.MinTurnaround = min
		stats.MaxTurnaround = max
	}

	return stats, nil
}

// GetRejectionRates returns rejection rates grouped by type and reason
func GetRejectionRates(modelType string) (RejectionRate, error) {
	db := getDB()
	tableName := getModelTableName(modelType)
	if tableName == "" {
		return RejectionRate{}, fmt.Errorf("invalid model type: %s", modelType)
	}

	var rejectionRate RejectionRate
	rejectionRate.ModelType = modelType

	// Get total created and total rejected - use model-specific queries
	var totalCreated int64
	var totalRejected int64

	switch modelType {
	case "profile":
		db.Model(&models.Profile{}).Count(&totalCreated)
		db.Model(&models.Profile{}).Where("status = ?", "inactive").Count(&totalRejected)
	case "business":
		db.Model(&models.Business{}).Count(&totalCreated)
		db.Model(&models.Business{}).Where("status = ?", "Rejected").Count(&totalRejected)
	case "event":
		db.Model(&models.Event{}).Count(&totalCreated)
		db.Model(&models.Event{}).Where("status = ?", "Rejected").Count(&totalRejected)
	case "achievement":
		db.Model(&models.Achievement{}).Count(&totalCreated)
		db.Model(&models.Achievement{}).Where("status = ?", "Rejected").Count(&totalRejected)
	}

	rejectionRate.TotalCreated = totalCreated
	rejectionRate.TotalRejected = totalRejected

	if totalCreated > 0 {
		rejectionRate.RejectionRate = float64(totalRejected) / float64(totalCreated) * 100
	}

	// Get rejections grouped by reason (from rejected_at field context)
	// We'll need to look at the actual data - for now return empty by reason
	// In a real implementation, you'd store rejection reason in a separate field
	var byReason []RejectionByReason
	rejectionRate.ByReason = byReason

	return rejectionRate, nil
}

// GetAllApprovalTurnarounds returns turnaround stats for all model types
func GetAllApprovalTurnarounds() ([]TurnaroundStats, error) {
	var allStats []TurnaroundStats

	for _, modelType := range []string{"profile", "business", "event", "achievement"} {
		stats, err := GetApprovalTurnaround(modelType)
		if err != nil {
			return nil, err
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

// GetAllRejectionRates returns rejection rates for all model types
func GetAllRejectionRates() ([]RejectionRate, error) {
	var allRates []RejectionRate

	for _, modelType := range []string{"profile", "business", "event", "achievement"} {
		rate, err := GetRejectionRates(modelType)
		if err != nil {
			return nil, err
		}
		allRates = append(allRates, rate)
	}

	return allRates, nil
}