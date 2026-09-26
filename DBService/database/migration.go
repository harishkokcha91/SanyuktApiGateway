package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// getMigrationsPath returns the migrations path from environment variable or computes a sensible default
func getMigrationsPath() string {
	// Check environment variable first (for Docker, AWS, CI/CD)
	if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
		// Ensure it's a valid file:// URL
		if filepath.IsAbs(envPath) {
			return "file://" + filepath.ToSlash(envPath)
		}
		return envPath
	}

	// Local development default: relative to project root
	// Try to find migrations directory from current working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("⚠️  Could not get working directory: %v", err)
	} else {
		// Look for migrations in common locations relative to wd
		candidates := []string{
			filepath.Join(wd, "..", "..", "migrations"),     // from DBService/
			filepath.Join(wd, "..", "migrations"),           // from DBService/cmd/migrate/
			filepath.Join(wd, "migrations"),                 // from project root
			"/app/migrations",                               // Docker default
			"/migrations",                                   // Alternative Docker
		}
		for _, candidate := range candidates {
			absPath, _ := filepath.Abs(candidate)
			if _, err := os.Stat(absPath); err == nil {
				return "file://" + filepath.ToSlash(absPath)
			}
		}
	}

	// Fallback: log warning and return empty (will fail with clear error)
	log.Println("⚠️  MIGRATIONS_PATH not set and migrations directory not found in standard locations")
	return ""
}

func makeMigrateInstance(dsn string) (*migrate.Migrate, error) {
	migrationsPath := getMigrationsPath()
	if migrationsPath == "" {
		return nil, fmt.Errorf("migrations path not configured: set MIGRATIONS_PATH environment variable")
	}
	log.Printf("🔍 Using migrations path: %s", migrationsPath)
	return migrate.New(migrationsPath, dsn)
}

// RunMigrations runs all pending database migrations using golang-migrate
func RunMigrations() {
	loadEnv()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "6501")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "admin")
	dbname := getEnv("DB_NAME", "sanyukt")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	log.Println("🚀 Starting database migration...")

	m, err := makeMigrateInstance(dsn)
	if err != nil {
		log.Fatalf("❌ Failed to create migrate instance: %v", err)
	}

	// Run all pending migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("✅ No pending migrations")
	} else {
		log.Println("✅ Database migrated successfully")
	}

	// Print current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("⚠️  Could not get migration version: %v", err)
	} else if err == migrate.ErrNilVersion {
		log.Println("📋 Migration version: none (fresh database)")
	} else {
		log.Printf("📋 Migration version: %d (dirty: %v)", version, dirty)
	}
}

// RollbackMigration rolls back the last migration (use with caution in production)
func RollbackMigration() {
	loadEnv()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "6501")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "admin")
	dbname := getEnv("DB_NAME", "sanyukt")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	log.Println("🔄 Rolling back last migration...")

	m, err := makeMigrateInstance(dsn)
	if err != nil {
		log.Fatalf("❌ Failed to create migrate instance: %v", err)
	}

	if err := m.Steps(-1); err != nil {
		log.Fatalf("❌ Rollback failed: %v", err)
	}
	log.Println("✅ Rolled back one migration")

	version, dirty, _ := m.Version()
	log.Printf("📋 Current migration version: %d (dirty: %v)", version, dirty)
}

// MigrationStatus prints the current migration status
func MigrationStatus() {
	loadEnv()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "6501")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "admin")
	dbname := getEnv("DB_NAME", "sanyukt")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	m, err := makeMigrateInstance(dsn)
	if err != nil {
		log.Fatalf("❌ Failed to create migrate instance: %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatalf("❌ Failed to get migration version: %v", err)
	}

	if err == migrate.ErrNilVersion {
		fmt.Println("No migrations applied yet")
		return
	}

	fmt.Printf("Current version: %d\n", version)
	fmt.Printf("Dirty: %v\n", dirty)

	// List all available migrations
	migrationsDir := getMigrationsDir()
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Printf("⚠️  Could not read migrations directory: %v", err)
		return
	}

	fmt.Println("\nAvailable migrations:")
	for _, f := range files {
		if !f.IsDir() && (len(f.Name()) > 3 && (f.Name()[:3] >= "000" && f.Name()[:3] <= "999")) {
			fmt.Printf("  %s\n", f.Name())
		}
	}
}

func getMigrationsDir() string {
	if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
		return envPath
	}
	wd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(wd, "..", "..", "migrations"),
		filepath.Join(wd, "..", "migrations"),
		filepath.Join(wd, "migrations"),
		"/app/migrations",
		"/migrations",
	}
	for _, candidate := range candidates {
		absPath, _ := filepath.Abs(candidate)
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}
	return ""
}