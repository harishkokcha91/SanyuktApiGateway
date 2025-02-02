package main

import (
	"db-service/database"
)

func main() {
	// config.LoadEnv()
	// database.CreateDatabase()
	database.ConnectDB()

	database.RunMigrations()
	database.PrintAllTables()

}
