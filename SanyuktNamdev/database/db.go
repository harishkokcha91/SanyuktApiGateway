package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	// Load environment variables
	_ = godotenv.Load()

	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")
	sslmode := os.Getenv("PGSSLMODE")
	if sslmode == "" {
		sslmode = "require" // Production default
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
	log.Printf("Connecting to DB at %s:%s as %s (password redacted, sslmode=%s)", host, port, user, sslmode)

	// Configure GORM with silent logger to remove slow query logs
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Setup connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get sql.DB from GORM: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Ping to verify connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Database not reachable: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	DB = db
}

// WithContext returns a new DB instance with the given context for query tracing
func WithContext(ctx context.Context) *gorm.DB {
	return DB.WithContext(ctx)
}

// Close closes the database connection pool
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
