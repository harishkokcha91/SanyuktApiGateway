package database

import (
	"db-service/models"
	"fmt"
	"log"
)

func RunMigrations() {
	err := DB.AutoMigrate(&models.User{})
	if err != nil {
		fmt.Println("❌ Migration failed:", err)
		return
	}
	fmt.Println("✅ Database migrated successfully")
}

// Function to print all tables in the database
func PrintAllTables() {
	var tables []string
	rows, err := DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema='public'").Rows()
	if err != nil {
		log.Fatal("Error querying tables: ", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Fatal("Error scanning row: ", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Error after iteration: ", err)
	}

	fmt.Println("Tables in the database:")
	for _, table := range tables {
		fmt.Println(table)
	}
}
