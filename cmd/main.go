package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/ClickHouse/clickhouse-go"
)

type DepthOrder struct {
	Price   float64
	BaseQty float64
}

type OrderBook struct {
	ID       int64
	Exchange string
	Pair     string
	Asks     []DepthOrder
	Bids     []DepthOrder
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	// Connecting to ClickHouse
	db, err := sql.Open("clickhouse", fmt.Sprintf("tcp://%s:%s?username=%s&password=%s", dbHost, dbPort, dbUser, dbPassword))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Chek connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := runMigrations(db, "./migrations"); err != nil {
		log.Fatal(err)
	}
}

func runMigrations(db *sql.DB, migrationsPath string) error {
	return filepath.Walk(migrationsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".sql" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			queries := string(content)
			if _, err := db.Exec(queries); err != nil {
				return fmt.Errorf("error running migration %s: %w", path, err)
			}

			log.Printf("Applied migration: %s", path)
		}

		return nil
	})
}
