package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func loadEnv() {
	_ = godotenv.Load()
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func ConnectDB() {
	loadEnv()

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "6501")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "admin")
	dbname := getEnv("DB_NAME", "sanyukt")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	log.Printf("Connecting to DB at %s:%s as %s", host, port, user)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Connected to the database")
}

// CreateDatabase checks if the database exists, if not, it creates it.
func CreateDatabase() {
	loadEnv()

	host := getEnv("DB_HOST", "localhost")
	portStr := getEnv("DB_PORT", "6501")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "admin")
	dbname := getEnv("DB_NAME", "sanyukt")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port: %v", err)
	}

	// Connection string without the database name
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, user, password)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	fmt.Println("Successfully connected to the PostgreSQL server!")

	// Create a new database if it doesn't exist
	createDBSQL := fmt.Sprintf("CREATE DATABASE %s;", dbname)
	_, err = db.Exec(createDBSQL)
	if err != nil && !isDuplicateDatabaseError(err) {
		log.Fatalf("Error creating database: %v", err)
	}

	fmt.Printf("Successfully created or verified the database %s!\n", dbname)
}

// isDuplicateDatabaseError checks if the error is about the database already existing
func isDuplicateDatabaseError(err error) bool {
	return err != nil && ( // PostgreSQL error code 42P04: duplicate_database
	// Check for pq error code if using github.com/lib/pq
	err.Error() == fmt.Sprintf("pq: database \"%s\" already exists", os.Getenv("DB_NAME")))
}
