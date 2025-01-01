package bot

import (
	"encoding/json"
	"fmt"
	"github.com/centrifugal/centrifuge-go"
	"strconv"
	"strings"
)

// WebSocketHandler connects to the Nobitex WS, subscribes to orderbook updates.
func (bot *TradingBot) WebSocketHandler() {
	url := "wss://wss.nobitex.ir/connection/websocket"
	client := centrifuge.NewJsonClient(url, centrifuge.Config{})

	client.OnConnected(func(_ centrifuge.ConnectedEvent) {
		bot.openLogger.Info("Connected to WebSocket!")
	})
	client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
		bot.openLogger.WithField("reason", e.Reason).
			Warn("Disconnected from WebSocket")
	})

	channel := fmt.Sprintf("public:orderbook-%s", strings.ToUpper(bot.currencyPair))
	sub, err := client.NewSubscription(channel)
	if err != nil {
		bot.openLogger.WithError(err).Fatal("Failed to create subscription")
	}

	sub.OnPublication(func(event centrifuge.PublicationEvent) {
		if err := json.Unmarshal(event.Data, &bot.orderBook); err != nil {
			bot.openLogger.WithError(err).Error("Error parsing orderBook data")
			return
		}
		parsedAsks := bot.parseOrderBook(bot.orderBook.Asks)
		parsedBids := bot.parseOrderBook(bot.orderBook.Bids)

		bot.bookMutex.Lock()
		bot.orderBookGlobal.Asks = parsedAsks
		bot.orderBookGlobal.Bids = parsedBids
		bot.bookMutex.Unlock()

		bot.priceMu.Lock()
		if len(bot.orderBook.Asks) > 0 {
			bot.askBest, _ = strconv.ParseFloat(bot.orderBook.Asks[0][0], 64)
		}
		if len(bot.orderBook.Bids) > 0 {
			bot.bidBest, _ = strconv.ParseFloat(bot.orderBook.Bids[0][0], 64)
		}
		bot.priceMu.Unlock()
	})

	if err := sub.Subscribe(); err != nil {
		bot.openLogger.WithError(err).Fatal("Failed to subscribe to WS channel")
	}
	if err := client.Connect(); err != nil {
		bot.openLogger.WithError(err).Fatal("Failed to connect to WS")
	}
}
