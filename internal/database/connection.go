package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// Init initializes the database connection
func Init() error {
	dsn := getDatabaseDSN()

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		if closeErr := DB.Close(); closeErr != nil {
			log.Printf("Error closing database: %v", closeErr)
		}
		return err
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// getDatabaseDSN returns the database connection string
func getDatabaseDSN() string {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "app:app@tcp(127.0.0.1:3306)/servicesdb?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_ai_ci"
	}
	return dsn
}
