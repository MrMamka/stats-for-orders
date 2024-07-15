package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"stats-for-orders/internal/server"
	"stats-for-orders/internal/storage"
)

// Start database and run server
func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	migraionsPath := os.Getenv("MIGRATIONS_PATH")

	serverPort := flag.Int("port", 8080, "Port of server.")
	flag.Parse()

	db, err := storage.NewDataBase(dbHost, dbPort, dbUser, dbPassword, migraionsPath)
	if err != nil {
		log.Fatal("error creating database: ", err)
	}
	defer db.Close()

	s := server.NewServer(db)
	s.RegisterAndRun(fmt.Sprintf(":%d", *serverPort))
}
