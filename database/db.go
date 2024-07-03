// database/db.go

package database

import (
    "database/sql"
    "fmt"
    _ "github.com/lib/pq"
    "inventory-management-system/config"
)

var DB *sql.DB

func ConnectDatabase() error {
    psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
        "password=%s dbname=%s sslmode=disable",
        config.AppConfig.DBHost,
        config.AppConfig.DBPort,
        config.AppConfig.DBUser,
        config.AppConfig.DBPassword,
        config.AppConfig.DBName)

    var err error
    DB, err = sql.Open("postgres", psqlInfo)
    if err != nil {
        return fmt.Errorf("error opening database: %w", err)
    }

    if err = DB.Ping(); err != nil {
        return fmt.Errorf("error connecting to database: %w", err)
    }

    fmt.Println("Successfully connected to database!")
    return nil
}

func CloseDatabase() {
    if DB != nil {
        DB.Close()
    }
}