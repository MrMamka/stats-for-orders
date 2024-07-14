package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"stats-for-orders/internal/server"
	"stats-for-orders/internal/storage"
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
	migraionsPath := os.Getenv("MIGRATIONS_PATH")

	db, err := storage.NewDataBase(dbHost, dbPort, dbUser, dbPassword, migraionsPath)
	if err != nil {
		log.Fatal("error creating database: ", err)
	}
	defer db.Close()

	serverPort := flag.Int("port", 8080, "Port of server.")
	flag.Parse()

	s := server.NewServer(db)
	s.RegisterAndRun(fmt.Sprintf(":%d", *serverPort))
}
