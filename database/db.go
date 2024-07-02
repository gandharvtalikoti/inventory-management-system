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
	DB = db
	log.Println("Successfully connected to the database!")
	createTables()
}

func createTables() {
	studentTable := `CREATE TABLE IF NOT EXISTS students (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        age INT NOT NULL,
        grade INT NOT NULL,
        department_id INT,
        FOREIGN KEY (department_id) REFERENCES departments(id)
    )`

	departmentTable := `CREATE TABLE IF NOT EXISTS departments (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL
    )`

	if _, err := DB.Exec(studentTable); err != nil {
		log.Fatalf("Failed to create studentTable: %v", err)
	}
	if _, err := DB.Exec(departmentTable); err != nil {
		log.Fatalf("Failed to create departmentTable: %v", err)
	}
}
