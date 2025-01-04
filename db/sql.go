package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func ConnectionString() string {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	SQL_URI := os.Getenv("SQL_URI")

	return SQL_URI
}

func connect() *sql.DB {
	connStr := ConnectionString()

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Unable to open DB: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("Cannot ping database because %s", err)
	}

	log.Println("Successfully connected to database and pinged it")

	return db
}

func createTable(db *sql.DB, tableName string) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public'
			AND table_name = $1
		);
	`, tableName).Scan(&exists)

	if err != nil {
		log.Fatalf("Error checking if table exists: %v", err)
	}

	if !exists {
		query := `
		CREATE TABLE ` + tableName + ` (
			user_id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			rooms TEXT[]
		);
		`

		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}

		log.Println("Table '" + tableName + "' created successfully.")
	} else {
		log.Println("Table '" + tableName + "' already exists.")
	}
}

func getInstance() *sql.DB {
	db := connect()

	// createTable(db, "users")

	return db
}

func TestSQLDbConnection() {
	db := getInstance()

	fmt.Println(db)
}
