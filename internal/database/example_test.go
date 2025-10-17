package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Sternrassler/EVE-SDE-Database-Builder/internal/database"
	"github.com/jmoiron/sqlx"
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

// ExampleBatchInsert demonstrates basic batch insert usage
func ExampleBatchInsert() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create a test table
	_, err = db.Exec(`
		CREATE TABLE invTypes (
			typeID INTEGER,
			typeName TEXT,
			groupID INTEGER
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare data for batch insert
	columns := []string{"typeID", "typeName", "groupID"}
	rows := [][]interface{}{
		{34, "Tritanium", 18},
		{35, "Pyerite", 18},
		{36, "Mexallon", 18},
		{37, "Isogen", 18},
		{38, "Nocxium", 18},
	}

	// Perform batch insert
	ctx := context.Background()
	err = database.BatchInsert(ctx, db, "invTypes", columns, rows, 1000)
	if err != nil {
		log.Fatal(err)
	}

	// Verify data was inserted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM invTypes").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inserted %d rows\n", count)
	// Output: Inserted 5 rows
}

// ExampleBatchInsertWithProgress demonstrates batch insert with progress reporting
func ExampleBatchInsertWithProgress() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create a test table
	_, err = db.Exec(`CREATE TABLE test_data (id INTEGER, value INTEGER)`)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare 2500 rows (will be split into 3 batches with batchSize=1000)
	columns := []string{"id", "value"}
	rows := make([][]interface{}, 2500)
	for i := 0; i < 2500; i++ {
		rows[i] = []interface{}{i + 1, i * 10}
	}

	// Progress callback
	progressCallback := func(current, total int) {
		fmt.Printf("Progress: %d/%d rows inserted\n", current, total)
	}

	// Perform batch insert with progress reporting
	ctx := context.Background()
	err = database.BatchInsertWithProgress(ctx, db, "test_data", columns, rows, 1000, progressCallback)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// Progress: 1000/2500 rows inserted
	// Progress: 2000/2500 rows inserted
	// Progress: 2500/2500 rows inserted
}

// ExampleWithTransaction demonstrates basic transaction usage
func ExampleWithTransaction() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
		CREATE TABLE roles (user_id INTEGER, role TEXT);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Execute transaction
	ctx := context.Background()
	err = database.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
		// Insert user
		_, err := tx.Exec("INSERT INTO users (id, name) VALUES (?, ?)", 1, "Alice")
		if err != nil {
			return err
		}

		// Insert role
		_, err = tx.Exec("INSERT INTO roles (user_id, role) VALUES (?, ?)", 1, "admin")
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	// Verify data was committed
	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id = 1").Scan(&name)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: %s\n", name)
	// Output: User: Alice
}

// ExampleWithTransaction_rollback demonstrates automatic rollback on error
func ExampleWithTransaction_rollback() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)
	if err != nil {
		log.Fatal(err)
	}

	// Execute transaction that fails
	ctx := context.Background()
	err = database.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
		// This insert succeeds
		_, err := tx.Exec("INSERT INTO users (id, name) VALUES (?, ?)", 1, "Alice")
		if err != nil {
			return err
		}

		// This insert fails due to constraint violation (NOT NULL)
		_, err = tx.Exec("INSERT INTO users (id, name) VALUES (?, ?)", 2, nil)
		return err
	})

	if err != nil {
		fmt.Println("Transaction failed and was rolled back")
	}

	// Verify no data was committed (rollback worked)
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Users in database: %d\n", count)
	// Output:
	// Transaction failed and was rolled back
	// Users in database: 0
}

// ExampleWithTransaction_withOptions demonstrates using transaction options
func ExampleWithTransaction_withOptions() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create and populate table
	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
		INSERT INTO users (id, name) VALUES (1, 'Alice'), (2, 'Bob');
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Execute read-only transaction with serializable isolation
	ctx := context.Background()
	var count int
	err = database.WithTransaction(ctx, db, func(tx *sqlx.Tx) error {
		return tx.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	}, database.WithReadOnly(), database.WithIsolationLevel(sql.LevelSerializable))

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User count: %d\n", count)
	// Output: User count: 2
}

// ExampleQueryRow demonstrates querying a single row into a struct
func ExampleQueryRow() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create and populate table
	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER, name TEXT, active BOOLEAN);
		INSERT INTO users (id, name, active) VALUES (1, 'Alice', 1), (2, 'Bob', 0);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Define a struct for the result
	type User struct {
		ID     int    `db:"id"`
		Name   string `db:"name"`
		Active bool   `db:"active"`
	}

	// Query a single user
	ctx := context.Background()
	user, err := database.QueryRow[User](ctx, db, "SELECT id, name, active FROM users WHERE id = ?", 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("User: %s (ID: %d, Active: %t)\n", user.Name, user.ID, user.Active)
	// Output: User: Alice (ID: 1, Active: true)
}

// ExampleQueryAll demonstrates querying multiple rows into a slice of structs
func ExampleQueryAll() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create and populate table
	_, err = db.Exec(`
		CREATE TABLE products (id INTEGER, name TEXT, price REAL);
		INSERT INTO products (id, name, price) VALUES 
			(1, 'Tritanium', 5.5),
			(2, 'Pyerite', 8.2),
			(3, 'Mexallon', 45.0);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Define a struct for the results
	type Product struct {
		ID    int     `db:"id"`
		Name  string  `db:"name"`
		Price float64 `db:"price"`
	}

	// Query all products with price > 6
	ctx := context.Background()
	products, err := database.QueryAll[Product](ctx, db, "SELECT id, name, price FROM products WHERE price > ? ORDER BY price", 6.0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d products:\n", len(products))
	for _, p := range products {
		fmt.Printf("- %s: %.2f ISK\n", p.Name, p.Price)
	}
	// Output:
	// Found 2 products:
	// - Pyerite: 8.20 ISK
	// - Mexallon: 45.00 ISK
}

// ExampleExists demonstrates checking for row existence
func ExampleExists() {
	// Create an in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// Create and populate table
	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, email TEXT UNIQUE);
		INSERT INTO users (id, email) VALUES (1, 'alice@example.com');
	`)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Check if email already exists
	emailExists, err := database.Exists(ctx, db, "SELECT 1 FROM users WHERE email = ?", "alice@example.com")
	if err != nil {
		log.Fatal(err)
	}

	if emailExists {
		fmt.Println("Email is already registered")
	}

	// Check if different email exists
	newEmailExists, err := database.Exists(ctx, db, "SELECT 1 FROM users WHERE email = ?", "bob@example.com")
	if err != nil {
		log.Fatal(err)
	}

	if !newEmailExists {
		fmt.Println("Email is available")
	}

	// Output:
	// Email is already registered
	// Email is available
}

// ExampleNewTestDB demonstrates using the testing utility
func ExampleNewTestDB() {
	// Note: This example uses a mock testing.T for demonstration
	// In real tests, use the actual *testing.T from your test function

	// This is a simplified demonstration - in actual test code:
	// func TestMyFeature(t *testing.T) {
	//     db := database.NewTestDB(t)
	//     // Database is automatically migrated and cleaned up
	// }

	// For demonstration, we'll create a regular in-memory database
	db, err := database.NewDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close(db)

	// In NewTestDB, migrations are automatically applied
	// Here we manually apply them for demonstration
	err = database.ApplyMigrations(db)
	if err != nil {
		log.Fatal(err)
	}

	// Now we can use the database with schema already set up
	// Check if invTypes table exists (created by migration 001)
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='invTypes'").Scan(&tableName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Table exists: %s\n", tableName)
	// Output: Table exists: invTypes
}
