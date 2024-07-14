package server

import (
	"net/http"
	"stats-for-orders/internal/storage"

	"github.com/gin-gonic/gin"
)

type Server struct {
	db  *storage.DataBase
	mux *gin.Engine
}

// Create router and register handlers
func NewServer(db *storage.DataBase) *Server {
	mux := gin.Default()

	return &Server{
		db:  db,
		mux: mux,
	}
}

// Register handlers, then listen and serve
func (s *Server) RegisterAndRun(addr string) {
	s.registerHandlers()
	s.mux.Run(addr)
}

func (s *Server) registerHandlers() {
	s.mux.GET("/order-book", s.getOrderBook)
	s.mux.POST("/order-book", s.saveOrderBook)
	s.mux.GET("/order-history", s.getOrderHistory)
	s.mux.POST("/order", s.saveOrder)
}

func (s *Server) getOrderBook(c *gin.Context) {
	exchangeName := c.Query("exchange")
	pair := c.Query("pair")

	orderBooks, err := s.db.GetOrderBook(exchangeName, pair)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orderBooks)
}

func (s *Server) saveOrderBook(c *gin.Context) {
	var orderBook storage.OrderBook
	if err := c.BindJSON(&orderBook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.db.SaveOrderBook(&orderBook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "OrderBook saved successfully"})
}

func (s *Server) getOrderHistory(c *gin.Context) {

}

func (s *Server) saveOrder(c *gin.Context) {

}

// func GetOrderBook(exchange_name, pair string) ([]*OrderBook, error)
// func SaveOrderBook(exchange_name, pair string, orderBook []*OrderBook) error
// func GetOrderHistory(client *Client) ([]*HistoryOrder, error)
// func SaveOrder(client *Client, order *HistoryOrder) error
