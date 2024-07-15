package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"stats-for-orders/internal/storage"

	"github.com/gin-gonic/gin"
)

type Server struct {
	db  *storage.DataBase
	mux *gin.Engine
}

// Create router with middleware and register handlers
func NewServer(db *storage.DataBase) *Server {
	mux := gin.Default()
	mux.Use(logMiddleware)

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

func logMiddleware(c *gin.Context) {
	method := c.Request.Method
	path := c.Request.URL.Path
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	fmt.Printf("Got request. Method: %s; Path: %s; Body: %s\n", method, path, string(bodyBytes))

	c.Next()
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
	client := &storage.Client{
		ClientName:   c.Query("client_name"),
		ExchangeName: c.Query("exchange_name"),
		Label:        c.Query("label"),
		Pair:         c.Query("pair"),
	}

	historyOrders, err := s.db.GetOrderHistory(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, historyOrders)
}

func (s *Server) saveOrder(c *gin.Context) {
	var order storage.HistoryOrder
	if err := c.BindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.db.SaveOrder(&order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Order saved successfully"})
}
