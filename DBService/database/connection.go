package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
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

func ConnectDB() {
	var err error
	// dsn := os.Getenv("DB_URL")
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, user, password)
	log.Printf(dsn)
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to the database")

}

// CreateDatabase checks if the database exists, if not, it creates it.
func CreateDatabase() {

	// Connection string without the database name
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable",
		host, port, user, password)

	// Open a connection to the PostgreSQL server
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
	defer db.Close()

	// Ping the database to verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %q", err)
	}

	fmt.Println("Successfully connected to the PostgreSQL server!")

	// Create a new database
	createDBSQL := fmt.Sprintf("CREATE DATABASE %s;", dbname)
	_, err = db.Exec(createDBSQL)
	if err != nil {
		log.Fatalf("Error creating database: %q", err)
	}

	fmt.Printf("Successfully created the database %s!\n", db)
}
