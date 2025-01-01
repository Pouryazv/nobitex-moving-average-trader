package bot

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"log"
	"nobitex-sma-bot/internal/logs"
	"nobitex-sma-bot/internal/nobitex"
	"os"
	"sync"
	"time"
)

// ----------------------------------------------------------------------------
// Bot Struct
// ----------------------------------------------------------------------------

type (
	// OrderBook carries the raw (string) bids/asks from the WebSocket
	OrderBook struct {
		Asks [][]string `json:"asks"`
		Bids [][]string `json:"bids"`
	}

	// UpdatedOrderBook keeps numeric forms of bids/asks.
	UpdatedOrderBook struct {
		Asks [][2]float64
		Bids [][2]float64
	}
)
type TradingBot struct {
	apiToken     string
	currencyPair string

	// Logging
	openLogger  *logrus.Logger
	closeLogger *logrus.Logger

	// Position tracking
	positionCount      int
	balanceInPositions float64
	posMutex           sync.Mutex

	// WebSocket order book
	orderBook       OrderBook
	orderBookGlobal UpdatedOrderBook
	bookMutex       sync.Mutex

	// Best bid/ask
	bidBest float64
	askBest float64
	priceMu sync.RWMutex

	// OCO tracking
	ocoOrders map[int]bool
	ocoMu     sync.Mutex

	// Concurrency flags
	buyOrderRunning  bool
	buyOrderMu       sync.Mutex
	sellOrderRunning bool
	sellOrderMu      sync.Mutex
}

// ----------------------------------------------------------------------------
// Bot Constructor & Entry Point
// ----------------------------------------------------------------------------

// setupLoggers configures two file-based loggers: one for open positions, one for closed.
func (bot *TradingBot) setupLoggers() {
	// Where we want to store logs

	logDir := "log/" + bot.currencyPair

	// Create 'open positions' logger
	openLogger, err := logs.Filelogger(logDir, "positions_open.log", logrus.InfoLevel)
	if err != nil {

		log.Fatalf("Failed to create openLogger: %v", err)

	}

	// Create 'close positions' logger
	closeLogger, err := logs.Filelogger(logDir, "positions_close.log", logrus.InfoLevel)
	if err != nil {

		log.Fatalf("Failed to create closeLogger: %v", err)
	}

	bot.openLogger = openLogger
	bot.closeLogger = closeLogger
}

// NewTradingBot initializes the bot and sets up logging/env.
func NewTradingBot(pair string) *TradingBot {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	}
	apiToken := os.Getenv("NOBITEX_API_TOKEN")
	if apiToken == "" {
		log.Fatal("API token not found in environment variables")
	}

	bot := &TradingBot{
		apiToken:     apiToken,
		currencyPair: pair,
		ocoOrders:    make(map[int]bool),
	}
	bot.setupLoggers()
	return bot
}

// Run starts WebSocket subscription and enters the main trading loop.
func (bot *TradingBot) Run() {

	go bot.WebSocketHandler()
	time.Sleep(5 * time.Second) // Wait a bit for the order book to initialize

	for {
		bot.MonitorPositionsAndClose()
		prices, err := bot.fetchOHLCVData()
		if err != nil {
			bot.openLogger.WithError(err).Error("Error fetching OHLCV data")
			time.Sleep(5 * time.Second)
			continue
		}
		if len(prices) < Pastmin {
			bot.openLogger.Info("Not enough data for SMA. Retrying...")
			time.Sleep(5 * time.Second)
			continue
		}
		sma := bot.calculateSMA(prices) * 10
		balance, err := nobitex.GetAvailableBalance(bot.apiToken) // from your refactored code
		if err != nil {
			bot.openLogger.WithError(err).Error("Error fetching balance")
			time.Sleep(5 * time.Second)
			continue
		}

		bot.priceMu.RLock()
		bidBest := bot.bidBest
		askBest := bot.askBest
		bot.priceMu.RUnlock()

		bot.openLogger.WithFields(logrus.Fields{
			"bidBest": bidBest,
			"askBest": askBest,
			"SMA":     sma,
		}).Info("Current price and SMA")

		bot.posMutex.Lock()
		switch {
		case bidBest <= sma*(1-PriceDeviation) && bot.balanceInPositions < MinBalance && balance > MinBalance:
			bot.openLogger.WithFields(logrus.Fields{
				"balance":       MinBalance - bot.balanceInPositions,
				"price":         bidBest,
				"position_side": "buy(long)",
				"reason":        "price_below_sma_threshold",
			}).Info("Opening BUY position")

			bot.buyOrderMu.Lock()
			if !bot.buyOrderRunning {
				bot.buyOrderRunning = true
				go func() {
					defer func() {
						bot.buyOrderMu.Lock()
						bot.buyOrderRunning = false
						bot.buyOrderMu.Unlock()
					}()
					bot.PlaceBuyOrder(MinBalance-bot.balanceInPositions, bidBest)
				}()
			} else {
				bot.openLogger.Warn("BuyOrder thread already running.")
			}
			bot.buyOrderMu.Unlock()

		case askBest >= sma*(1+PriceDeviation) && bot.balanceInPositions < MinBalance && balance > MinBalance:
			bot.openLogger.WithFields(logrus.Fields{
				"balance":       MinBalance - bot.balanceInPositions,
				"price":         askBest,
				"position_side": "sell(short)",
				"reason":        "price_above_sma_threshold",
			}).Info("Opening SELL position")

			bot.sellOrderMu.Lock()
			if !bot.sellOrderRunning {
				bot.sellOrderRunning = true
				go func() {
					defer func() {
						bot.sellOrderMu.Lock()
						bot.sellOrderRunning = false
						bot.sellOrderMu.Unlock()
					}()
					bot.PlaceSellOrder(MinBalance-bot.balanceInPositions, askBest)
				}()
			} else {
				bot.openLogger.Warn("SellOrder thread already running.")
			}
			bot.sellOrderMu.Unlock()
		}
		bot.posMutex.Unlock()

		time.Sleep(5 * time.Second)
	}
}
