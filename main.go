package main

import (
	"inventory-management-system/config"
	"inventory-management-system/database"
    "inventory-management-system/routes"
	
)

func main() {
    // Load the configuration
    config.LoadConfig()

    // Connect to the database
    database.ConnectDatabase()
       r := routes.SetupRouter()
   r.Run()

}
