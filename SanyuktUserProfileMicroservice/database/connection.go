package database

import (
	"fmt"
	"log"
	"userprofile-service/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

const (
	host     = "localhost"
	port     = 6501
	user     = "postgres"
	password = "admin"
	dbname   = "sanyukt"
)

func Connect() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, user, password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// Auto-Migrate Models
	db.AutoMigrate(&models.UserProfile{})

	DB = db
	log.Println("Database connected successfully")
}
