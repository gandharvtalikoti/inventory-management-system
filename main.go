package main

import (
	"inventory-management-system/config"
	"inventory-management-system/database"
	
)

func main() {
    // Load the configuration
    config.LoadConfig()

    // Connect to the database
    database.ConnectDatabase()

}
