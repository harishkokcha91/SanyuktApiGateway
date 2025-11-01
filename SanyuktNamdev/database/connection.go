package database

import (
	"SanyuktNamdev/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {

	_ = godotenv.Load()

	// host := os.Getenv("DB_HOST")
	// port := os.Getenv("DB_PORT")
	// user := os.Getenv("DB_USER")
	// password := os.Getenv("DB_PASSWORD")

	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	// dbname := os.Getenv("PGDATABASE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=require",
		host, port, user, password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// Auto-Migrate Models
	db.AutoMigrate(&models.Achievement{})
	// Auto-Migrate Models
	db.AutoMigrate(&models.Business{})
	db.AutoMigrate(&models.Event{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Profile{})
	DB = db
	log.Println("Database connected successfully")
}
