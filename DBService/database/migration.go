package database

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs all pending database migrations using golang-migrate
func RunMigrations() {
	loadEnv()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "6501")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "admin")
	dbname := getEnv("DB_NAME", "sanyukt")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	// Migration files path (relative to project root)
	migrationsPath := "file://../../migrations"

	m, err := migrate.New(migrationsPath, dsn)
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
	migrationsPath := "file://../../migrations"

	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		log.Fatalf("❌ Failed to create migrate instance: %v", err)
	}

	if err := m.Steps(-1); err != nil {
		log.Fatalf("❌ Rollback failed: %v", err)
	}
	log.Println("✅ Rolled back one migration")
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
	migrationsPath := "file://../../migrations"

	m, err := migrate.New(migrationsPath, dsn)
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
	// Note: golang-migrate doesn't have a built-in way to list all migrations,
	// but we can read the directory
	files, err := os.ReadDir("../../migrations")
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