package main

import (
	"db-service/database"
	"flag"
	"fmt"
	"os"
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
		database.ConnectDB()

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
		database.ConnectDB()
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
	fmt.Println("  db-service init                    # Create database and run all migrations")
	fmt.Println("  db-service migrate                 # Run pending migrations (default)")
	fmt.Println("  db-service migrate -up             # Run pending migrations")
	fmt.Println("  db-service migrate -down           # Rollback last migration")
	fmt.Println("  db-service migrate -status         # Show migration status")
	fmt.Println("  db-service migrate -create <name>  # Create new migration files")
	fmt.Println("  db-service tables                  # List all tables in database")
}

func createMigration(name string) {
	// This is a helper to create new migration files
	// In practice, you'd create them manually in the migrations/ directory
	fmt.Printf("Create migration: %s\n", name)
	fmt.Println("Please create the following files manually in the migrations/ directory:")
	fmt.Printf("  %03d_%s.up.sql\n", getNextMigrationNumber()+1, name)
	fmt.Printf("  %03d_%s.down.sql\n", getNextMigrationNumber()+1, name)
}

func getNextMigrationNumber() int {
	// Simple helper to get next migration number
	// In practice, you'd read the migrations directory
	return 2 // We have 001 and 002 currently
}