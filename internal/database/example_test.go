package database_test

import (
	"fmt"
	"log"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
)

// ExampleNewDB demonstrates creating a new database connection
func ExampleNewDB() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert data
	_, err = db.Exec("INSERT INTO users (name) VALUES (?)", "Alice")
	if err != nil {
		log.Fatal(err)
	}

	// Query data
	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(name)
	// Output: Alice
}

// ExampleNewDB_fileDatabase demonstrates creating a file-based database
func ExampleNewDB_fileDatabase() {
	// Note: In production, use a proper path like "/var/lib/app/data.db"
	// For this example, we use :memory: to avoid cleanup
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Verify connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database connected")
	// Output: Database connected
}
