package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func connectDB() *sql.DB {
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=location sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("db init connection error: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("db connect error: %v", err)
	}

	fmt.Println("successful db connecting")
	return db
}
