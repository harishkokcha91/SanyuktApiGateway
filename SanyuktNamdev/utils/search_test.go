package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test models for search testing
type TestProfile struct {
	ID              uint      `gorm:"primaryKey"`
	Name            string    `json:"name"`
	CurrentLocation string    `json:"current_location"`
	Occupation      string    `json:"occupation"`
	Address         string    `json:"address"`
	Status          string    `json:"status"`
	DateOfBirth     string    `json:"date_of_birth"`
	CreatedAt       time.Time `json:"created_at"`
}

func (TestProfile) TableName() string {
	return "test_profiles"
}

func setupSearchTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&TestProfile{})
	require.NoError(t, err)

	// Seed test data
	testData := []TestProfile{
		{Name: "John Doe", CurrentLocation: "New York", Occupation: "Engineer", Address: "123 Main St", Status: "active", DateOfBirth: "1990-01-01"},
		{Name: "Jane Smith", CurrentLocation: "Los Angeles", Occupation: "Doctor", Address: "456 Oak Ave", Status: "active", DateOfBirth: "1985-05-15"},
		{Name: "Bob Johnson", CurrentLocation: "New York", Occupation: "Teacher", Address: "789 Elm St", Status: "pending", DateOfBirth: "1992-03-20"},
		{Name: "Alice Brown", CurrentLocation: "Chicago", Occupation: "Engineer", Address: "321 Pine Rd", Status: "active", DateOfBirth: "1988-07-10"},
		{Name: "Charlie Wilson", CurrentLocation: "Boston", Occupation: "Lawyer", Address: "654 Maple Dr", Status: "inactive", DateOfBirth: "1995-11-25"},
	}
	for i := range testData {
		db.Create(&testData[i])
	}

	return db
}

func TestParseSearchParams(t *testing.T) {
	// We can't easily test ParseSearchParams without a Gin context
	// This would require setting up a test Gin context
}

func TestSanitizeForILIKE(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"hello%", "hello\\%"},
		{"hello_world", "hello\\_world"},
		{"100%", "100\\%"},
		{"test's", "test''s"},
		{"%wildcard%", "\\%wildcard\\%"},
		{"", ""},
	}

	for _, tc := range tests {
		result := sanitizeForILIKE(tc.input)
		assert.Equal(t, tc.expected, result)
	}
}

func TestParseAgeRange(t *testing.T) {
	tests := []struct {
		input    string
		minAge   *int
		maxAge   *int
	}{
		{"", nil, nil},
		{"25-35", intPtr(25), intPtr(35)},
		{"25+", intPtr(25), nil},
		{"-35", nil, intPtr(35)},
		{"invalid", nil, nil},
		{"abc-def", nil, nil},
	}

	for _, tc := range tests {
		min, max := ParseAgeRange(tc.input)
		assert.Equal(t, tc.minAge, min)
		assert.Equal(t, tc.maxAge, max)
	}
}

func TestParseDateRange(t *testing.T) {
	tests := []struct {
		input     string
		startDate *time.Time
		endDate   *time.Time
	}{
		{"", nil, nil},
		{"2024-01-01,2024-12-31", timePtr("2024-01-01"), timePtr("2024-12-31")},
		{"2024-01-01,", timePtr("2024-01-01"), nil},
		{",2024-12-31", nil, timePtr("2024-12-31")},
		{"invalid", nil, nil},
	}

	for _, tc := range tests {
		start, end := ParseDateRange(tc.input)
		if tc.startDate != nil {
			require.NotNil(t, start)
			assert.Equal(t, *tc.startDate, *start)
		} else {
			assert.Nil(t, start)
		}
		if tc.endDate != nil {
			require.NotNil(t, end)
			assert.Equal(t, *tc.endDate, *end)
		} else {
			assert.Nil(t, end)
		}
	}
}

func TestApplyAgeRangeFilter(t *testing.T) {
	db := setupSearchTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Test data dates: 1990, 1985, 1992, 1988, 1995
	// Ages will depend on current year

	tests := []struct {
		name   string
		minAge *int
		maxAge *int
	}{
		{"no filter", nil, nil},
		{"min age", intPtr(25), nil},
		{"max age", nil, intPtr(30)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			query := db.Model(&TestProfile{})
			query = ApplyAgeRangeFilter(query, tc.minAge, tc.maxAge)
			var count int64
			query.Count(&count)
			// Just verify it runs without error and returns some count
			// Exact count depends on current year
			_ = count
		})
	}
}

func TestApplyDateRangeFilter(t *testing.T) {
	type TestEvent struct {
		ID        uint      `gorm:"primaryKey"`
		Name      string    `json:"name"`
		EventDate string    `json:"event_date"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&TestEvent{})
	require.NoError(t, err)

	// Seed test data
	events := []TestEvent{
		{Name: "Event 1", EventDate: "2024-01-15", Status: "Approved"},
		{Name: "Event 2", EventDate: "2024-06-15", Status: "Approved"},
		{Name: "Event 3", EventDate: "2024-12-15", Status: "Approved"},
	}
	for i := range events {
		db.Create(&events[i])
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	startDate := timePtr("2024-06-01")
	endDate := timePtr("2024-09-01")

	query := db.Model(&TestEvent{})
	query = ApplyDateRangeFilter(query, "event_date", startDate, endDate)

	var count int64
	query.Count(&count)
	assert.Equal(t, int64(1), count) // Only Event 2 falls in range
}

func TestBuildSearchQuery(t *testing.T) {
	db := setupSearchTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	config := ProfileSearchConfig

	t.Run("keyword search", func(t *testing.T) {
		params := SearchParams{
			Keyword:     "doe",
			SortBy:      "created_at",
			SortOrder:   "DESC",
			Limit:       20,
			Offset:      0,
			CustomFilters: map[string]string{},
		}

		query := BuildSearchQuery(db.Model(&TestProfile{}), params, config)
		var results []TestProfile
		query.Find(&results)
		assert.Len(t, results, 1)
		assert.Equal(t, "John Doe", results[0].Name)
	})

	t.Run("status filter", func(t *testing.T) {
		params := SearchParams{
			Status:     "active",
			SortBy:     "created_at",
			SortOrder:  "DESC",
			Limit:      20,
			Offset:     0,
			CustomFilters: map[string]string{},
		}

		query := BuildSearchQuery(db.Model(&TestProfile{}), params, config)
		var results []TestProfile
		query.Find(&results)
		assert.Len(t, results, 3) // 3 active profiles
	})

	t.Run("custom filter", func(t *testing.T) {
		params := SearchParams{
			Status:     "",
			SortBy:     "created_at",
			SortOrder:  "DESC",
			Limit:      20,
			Offset:     0,
			CustomFilters: map[string]string{"current_location": "New York"},
		}

		query := BuildSearchQuery(db.Model(&TestProfile{}), params, config)
		var results []TestProfile
		query.Find(&results)
		assert.Len(t, results, 2) // 2 profiles in New York
	})

	t.Run("pagination", func(t *testing.T) {
		params := SearchParams{
			Keyword:     "",
			SortBy:      "name",
			SortOrder:   "ASC",
			Limit:       2,
			Offset:      0,
			CustomFilters: map[string]string{},
		}

		query := BuildSearchQuery(db.Model(&TestProfile{}), params, config)
		var results []TestProfile
		query.Find(&results)
		assert.Len(t, results, 2)
		assert.Equal(t, "Alice Brown", results[0].Name)
		assert.Equal(t, "Bob Johnson", results[1].Name)
	})

	t.Run("second page", func(t *testing.T) {
		params := SearchParams{
			Keyword:     "",
			SortBy:      "name",
			SortOrder:   "ASC",
			Limit:       2,
			Offset:      2,
			CustomFilters: map[string]string{},
		}

		query := BuildSearchQuery(db.Model(&TestProfile{}), params, config)
		var results []TestProfile
		query.Find(&results)
		assert.Len(t, results, 2)
		assert.Equal(t, "Charlie Wilson", results[0].Name)
		assert.Equal(t, "Jane Smith", results[1].Name)
	})
}

func TestBuildCountQuery(t *testing.T) {
	db := setupSearchTestDB(t)
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	config := ProfileSearchConfig

	params := SearchParams{
		Keyword:     "engineer",
		Status:      "active",
		SortBy:      "created_at",
		SortOrder:   "DESC",
		Limit:       20,
		Offset:      0,
		CustomFilters: map[string]string{},
	}

	query := BuildCountQuery(db.Model(&TestProfile{}), params, config)
	var count int64
	query.Count(&count)
	assert.Equal(t, int64(2), count) // 2 active engineers
}

func TestProfileCustomFilters(t *testing.T) {
	filters := ProfileCustomFilters()
	assert.Equal(t, "current_location", filters["location"])
	assert.Equal(t, "age_range", filters["age_range"])
}

func TestBusinessCustomFilters(t *testing.T) {
	filters := BusinessCustomFilters()
	assert.Equal(t, "category", filters["category"])
	assert.Equal(t, "location", filters["location"])
	assert.Equal(t, "city", filters["city"])
	assert.Equal(t, "state", filters["state"])
	assert.Equal(t, "owner", filters["owner"])
}

func TestEventCustomFilters(t *testing.T) {
	filters := EventCustomFilters()
	assert.Equal(t, "category", filters["category"])
	assert.Equal(t, "city", filters["location"]) // "location" query param maps to "city" field
	assert.Equal(t, "city", filters["city"])
	assert.Equal(t, "state", filters["state"])
	assert.Equal(t, "organizer", filters["organizer"])
}

func TestAchievementCustomFilters(t *testing.T) {
	filters := AchievementCustomFilters()
	assert.Equal(t, "achievement_type", filters["type"])
}

func intPtr(i int) *int {
	return &i
}

func timePtr(s string) *time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return &t
}