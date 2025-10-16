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
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)

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

