// package main

// import (
// 	"database/sql"
// 	"fmt"
// 	"log"

// 	_ "github.com/lib/pq"
// )

// func createTable(db *sql.DB) {
// 	query := `
// 	CREATE TABLE IF NOT EXISTS services (
// 		id SERIAL PRIMARY KEY,
// 		name VARCHAR(255) NOT NULL,
// 		url VARCHAR(255) NOT NULL,
// 		healthy BOOLEAN NOT NULL,
// 		timestamp TIMESTAMP NOT NULL
// 	);
// 	`
// 	_, err := db.Exec(query)
// 	if err != nil {
// 		log.Fatal("Error creating table: ", err)
// 	}
// 	fmt.Println("Table created successfully!")
// }

// func main() {

// 	// PostgreSQL connection string
// 	connStr := "host=localhost port=5432 user=admin password=admin123 dbname=healthcheck sslmode=disable"

// 	// Open database connection
// 	db, err := sql.Open("postgres", connStr)
// 	if err != nil {
// 		log.Fatal("Error while opening connection: ", err)
// 	}

// 	// Check connection
// 	err = db.Ping()
// 	if err != nil {
// 		log.Fatal("Cannot connect to database: ", err)
// 	}

// 	fmt.Println("Connected to PostgreSQL successfully!")

// 	createTable(db)

// 	// Example query
// 	rows, err := db.Query("SELECT version()")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var version string
// 		rows.Scan(&version)
// 		fmt.Println("PostgreSQL Version:", version)
// 	}

// 	defer db.Close()
// }

package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost" // default for local development
	}

	connStr := fmt.Sprintf(
		"host=%s port=5432 user=admin password=admin123 dbname=healthcheck sslmode=disable",
		host,
	)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("Cannot reach database: ", err)
	}

	fmt.Println("✅ Connected to PostgreSQL")
	createTable()
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS services (
		id        SERIAL PRIMARY KEY,
		name      VARCHAR(100) UNIQUE,
		url       VARCHAR(255),
		healthy   BOOLEAN,
		timestamp VARCHAR(50)
	);`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}
	fmt.Println("✅ Table ready")
}
