package main

import (
	"flag"
	"fmt"
	"os"

	"db-service/database"
)

func main() {
	// Define CLI commands
	migrateCmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	migrateDown := migrateCmd.Bool("down", false, "Rollback one migration")
	migrateStatus := migrateCmd.Bool("status", false, "Show migration status")
	migrateCreate := migrateCmd.String("create", "", "Create new migration files (provide name)")

	// Parse top-level command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "migrate":
		migrateCmd.Parse(os.Args[2:])

		if *migrateCreate != "" {
			createMigration(*migrateCreate)
		} else if *migrateDown {
			database.RollbackMigration()
		} else if *migrateStatus {
			database.MigrationStatus()
		} else {
			// Default: run migrations up
			database.RunMigrations()
		}

	case "init":
		// Initialize database (create if not exists) then run migrations
		database.CreateDatabase()
		database.RunMigrations()
		database.PrintAllTables()

	case "tables":
		database.ConnectDB()
		database.PrintAllTables()

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("DBService - Database Management Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd/migrate init                    # Create database and run all migrations")
	fmt.Println("  go run ./cmd/migrate migrate                 # Run pending migrations (default)")
	fmt.Println("  go run ./cmd/migrate migrate -down           # Rollback last migration")
	fmt.Println("  go run ./cmd/migrate migrate -status         # Show migration status")
	fmt.Println("  go run ./cmd/migrate migrate -create <name>  # Create new migration files")
	fmt.Println("  go run ./cmd/migrate tables                  # List all tables in database")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  DB_HOST         Database host (default: localhost)")
	fmt.Println("  DB_PORT         Database port (default: 6501)")
	fmt.Println("  DB_USER         Database user (default: postgres)")
	fmt.Println("  DB_PASSWORD     Database password (default: admin)")
	fmt.Println("  DB_NAME         Database name (default: sanyukt)")
	fmt.Println("  MIGRATIONS_PATH Path to migrations directory (auto-detected if not set)")
}

func createMigration(name string) {
	// This is a helper to create new migration files
	// In practice, you'd create them manually in the migrations/ directory
	fmt.Printf("Create migration: %s\n", name)
	fmt.Println("Please create the following files manually in the migrations/ directory:")
	fmt.Printf("  00X_%s.up.sql\n", name)
	fmt.Printf("  00X_%s.down.sql\n", name)
}