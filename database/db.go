package database

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDatabase() {
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL is not set in .env file")
    }

    dsn := fmt.Sprintf("%s@tcp(localhost:3306)/inventory-sys", dbURL)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("Error connecting to database: %v", err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatalf("Error pinging database: %v", err)
    }
    DB = db
    log.Println("Successfully connected to the database!")
}
