package tests

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"stats-for-orders/internal/storage"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

var testMan *testManager

type Generator struct {
	*rand.Rand
}

func NewGenerator() Generator {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	return Generator{rand.New(source)}
}

func (g Generator) GenerateInt() int {
	return g.Intn(math.MaxInt32)
}

func (g Generator) GenerateString() string {
	return fmt.Sprintf("String%d", g.GenerateInt())
}

func (g Generator) GenerateFloat() float64 {
	return g.Float64() * 100
}

func (g Generator) GenerateDepthOrders() []storage.DepthOrder {
	size := g.Intn(10)
	depthOrders := make([]storage.DepthOrder, size)
	for i := range depthOrders {
		depthOrders[i].Price = g.GenerateFloat()
		depthOrders[i].BaseQty = g.GenerateFloat()
	}

	return depthOrders
}

type testManager struct {
	addr    string
	client  *resty.Client
	randGen Generator
}

func newTestManager(addr string) *testManager {
	return &testManager{addr: addr, client: resty.New(), randGen: NewGenerator()}
}

func (m *testManager) sendPostRequest(endpoint string, data interface{}) (*resty.Response, error) {
	resp, err := m.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post(m.addr + endpoint)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *testManager) sendGetRequest(endpoint string) (*resty.Response, error) {
	resp, err := m.client.R().
		Get(m.addr + endpoint)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func init() {
	testMan = newTestManager("http://localhost:8080")
}

// Before run this tests you have to up service (docker compose up)
func TestSaveAndGetOrderBook(t *testing.T) {
	reqOrderBook := storage.OrderBook{
		ID:       int64(testMan.randGen.GenerateInt()),
		Exchange: testMan.randGen.GenerateString(),
		Pair:     testMan.randGen.GenerateString(),
		Asks:     testMan.randGen.GenerateDepthOrders(),
		Bids:     testMan.randGen.GenerateDepthOrders(),
	}

	resp, err := testMan.sendPostRequest("/order-book", reqOrderBook)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())

	resp, err = testMan.sendGetRequest(fmt.Sprintf("/order-book?exchange=%s&pair=%s", reqOrderBook.Exchange, reqOrderBook.Pair))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	var respOrderBooks []storage.OrderBook
	err = json.Unmarshal(resp.Body(), &respOrderBooks)
	require.NoError(t, err)
	require.Len(t, respOrderBooks, 1)
	require.Equal(t, reqOrderBook, respOrderBooks[0])
}

func TestSaveAndGetOrderHistory(t *testing.T) {
	reqOrder := storage.HistoryOrder{
		Client: &storage.Client{
			ClientName:   testMan.randGen.GenerateString(),
			ExchangeName: testMan.randGen.GenerateString(),
			Label:        testMan.randGen.GenerateString(),
			Pair:         testMan.randGen.GenerateString(),
		},
		Side:                testMan.randGen.GenerateString(),
		Type:                testMan.randGen.GenerateString(),
		BaseQty:             testMan.randGen.GenerateFloat(),
		Price:               testMan.randGen.GenerateFloat(),
		AlgorithmNamePlaced: testMan.randGen.GenerateString(),
		LowestSellPrc:       testMan.randGen.GenerateFloat(),
		HighestBuyPrc:       testMan.randGen.GenerateFloat(),
		CommissionQuoteQty:  testMan.randGen.GenerateFloat(),
		TimePlaced:          time.Now(),
	}

	resp, err := testMan.sendPostRequest("/order", reqOrder)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())

	resp, err = testMan.sendGetRequest(fmt.Sprintf(
		"/order-history?client_name=%s&exchange_name=%s&label=%s&pair=%s",
		reqOrder.Client.ClientName, reqOrder.Client.ExchangeName, reqOrder.Client.Label, reqOrder.Client.Pair))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	var respOrders []storage.HistoryOrder
	err = json.Unmarshal(resp.Body(), &respOrders)
	require.NoError(t, err)
	require.Len(t, respOrders, 1)

	require.Equal(t, reqOrder.TimePlaced.Second(), respOrders[0].TimePlaced.Second())
	respOrders[0].TimePlaced = reqOrder.TimePlaced
	require.Equal(t, reqOrder, respOrders[0])
}
