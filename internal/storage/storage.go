package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/ClickHouse/clickhouse-go"
)

type Client struct {
	ClientName   string `json:"client_name"`
	ExchangeName string `json:"exchange_name"`
	Label        string `json:"label"`
	Pair         string `json:"pair"`
}

type DepthOrder struct {
	Price   float64 `json:"price"`
	BaseQty float64 `json:"base_qty"`
}

type OrderBook struct {
	ID       int64        `json:"id"`
	Exchange string       `json:"exchange"`
	Pair     string       `json:"pair"`
	Asks     []DepthOrder `json:"asks"`
	Bids     []DepthOrder `json:"bids"`
}

type HistoryOrder struct {
	Client              *Client   `json:"client"`
	Side                string    `json:"side"`
	Type                string    `json:"type"`
	BaseQty             float64   `json:"base_qty"`
	Price               float64   `json:"price"`
	AlgorithmNamePlaced string    `json:"algorithm_name_placed"`
	LowestSellPrc       float64   `json:"lowest_sell_prc"`
	HighestBuyPrc       float64   `json:"highest_buy_prc"`
	CommissionQuoteQty  float64   `json:"commission_quote_qty"`
	TimePlaced          time.Time `json:"time_placed"`
}

type DataBase struct {
	db *sql.DB
}

// Connects to database and migrate
func NewDataBase(host, port, user, password, migrationsPath string) (*DataBase, error) {
	// Connecting to ClickHouse
	db, err := sql.Open("clickhouse", fmt.Sprintf("tcp://%s:%s?username=%s&password=%s", host, port, user, password))
	if err != nil {
		return nil, err
	}

	// Chek connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := runMigrations(db, migrationsPath); err != nil {
		return nil, err
	}

	return &DataBase{db: db}, nil
}

func (db *DataBase) Close() {
	db.db.Close()
}

// Walk through all .sql files in directory and exec queries in them
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

func arraysToDepthOrder(arrays [][]float64) ([]DepthOrder, error) {
	depthOrders := make([]DepthOrder, len(arrays))
	for i, array := range arrays {
		if len(array) != 2 {
			return nil, fmt.Errorf("invalid array length: %d", len(array))
		}
		depthOrders[i] = DepthOrder{
			Price:   array[0],
			BaseQty: array[1],
		}
	}
	return depthOrders, nil
}

func (db *DataBase) GetOrderBook(exchangeName, pair string) ([]*OrderBook, error) {
	query := `SELECT id, asks, bids FROM OrderBook WHERE exchange = ? AND pair = ?`
	rows, err := db.db.Query(query, exchangeName, pair)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orderBooks []*OrderBook
	for rows.Next() {
		var (
			id   int64
			asks [][]float64
			bids [][]float64
		)
		if err := rows.Scan(&id, &asks, &bids); err != nil {
			return nil, err
		}

		asksDepthOrder, err := arraysToDepthOrder(asks)
		if err != nil {
			return nil, err
		}
		bidsDepthOrder, err := arraysToDepthOrder(bids)
		if err != nil {
			return nil, err
		}

		orderBook := &OrderBook{
			ID:       id,
			Exchange: exchangeName,
			Pair:     pair,
			Asks:     asksDepthOrder,
			Bids:     bidsDepthOrder,
		}
		orderBooks = append(orderBooks, orderBook)
	}

	return orderBooks, nil
}

func depthOrderToArrays(depthOrder []DepthOrder) [][]float64 {
	arrays := make([][]float64, len(depthOrder))
	for i, order := range depthOrder {
		arrays[i] = []float64{order.Price, order.BaseQty}
	}
	return arrays
}

func (db *DataBase) SaveOrderBook(orderBook *OrderBook) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO OrderBook (id, exchange, pair, asks, bids) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	askArrays := depthOrderToArrays(orderBook.Asks)
	bidArrays := depthOrderToArrays(orderBook.Bids)

	if _, err := stmt.Exec(
		orderBook.ID,
		orderBook.Exchange,
		orderBook.Pair,
		askArrays,
		bidArrays,
	); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *DataBase) GetOrderHistory(client *Client) ([]*HistoryOrder, error) {
	query := `SELECT side, type, base_qty, price, algorithm_name_placed, lowest_sell_prc,
	highest_buy_prc, commission_quote_qty, time_placed 
	FROM Order_History WHERE client_name = ? AND exchange_name = ? AND label = ? AND pair = ?`

	rows, err := db.db.Query(query, client.ClientName, client.ExchangeName, client.Label, client.Pair)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var historyOrders []*HistoryOrder
	for rows.Next() {
		var historyOrder HistoryOrder
		if err := rows.Scan(
			&historyOrder.Side,
			&historyOrder.Type,
			&historyOrder.BaseQty,
			&historyOrder.Price,
			&historyOrder.AlgorithmNamePlaced,
			&historyOrder.LowestSellPrc,
			&historyOrder.HighestBuyPrc,
			&historyOrder.CommissionQuoteQty,
			&historyOrder.TimePlaced,
		); err != nil {
			return nil, err
		}
		historyOrder.Client = client
		historyOrders = append(historyOrders, &historyOrder)
	}

	return historyOrders, nil
}

func (db *DataBase) SaveOrder(order *HistoryOrder) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`INSERT INTO Order_History 
	(client_name, exchange_name, label, pair, side, type, base_qty, price,
	algorithm_name_placed, lowest_sell_prc, highest_buy_prc, commission_quote_qty, time_placed)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)

	if err != nil {
		return err
	}
	defer stmt.Close()

	if order.Client == nil {
		order.Client = &Client{}
	}

	if _, err := stmt.Exec(
		order.Client.ClientName,
		order.Client.ExchangeName,
		order.Client.Label,
		order.Client.Pair,
		order.Side,
		order.Type,
		order.BaseQty,
		order.Price,
		order.AlgorithmNamePlaced,
		order.LowestSellPrc,
		order.HighestBuyPrc,
		order.CommissionQuoteQty,
		order.TimePlaced,
	); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
