// main.go

package main

import (
    "inventory-management-system/config"
    "inventory-management-system/database"
    "log"
    //"inventory-management-system/routes"
)

func main() {
    // Load the configuration
    if err := config.LoadConfig(); err != nil {
        log.Fatalf("Error loading config: %v", err)
    }

    // Connect to the database
    if err := database.ConnectDatabase(); err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }
    defer database.CloseDatabase()

    // r := routes.SetupRouter()
    // r.Run()

    // Your main application logic goes here
}