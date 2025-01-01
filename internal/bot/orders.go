package bot

import (
	"github.com/sirupsen/logrus"
	"nobitex-sma-bot/internal/nobitex"
	"time"
)

// ----------------------------------------------------------------------------
// Order Management (Place Buy/Sell in increments)
// ----------------------------------------------------------------------------

func (bot *TradingBot) PlaceBuyOrder(balance float64, maxPrice float64) {
	var (
		prevOrderID    int
		prevOrderPrice float64
		prevPrice      float64
		totalRemaining = balance
		cancelRetries  = 0
	)

	bot.openLogger.WithFields(logrus.Fields{
		"action":    "PlaceBuyOrder_start",
		"balance":   balance,
		"max_price": maxPrice,
	}).Info("Starting buy orders")

	for {
		bot.bookMutex.Lock()
		bids := bot.orderBookGlobal.Bids
		bot.bookMutex.Unlock()

		if len(bids) == 0 {
			bot.openLogger.Warn("No bid price available. Waiting...")
			time.Sleep(time.Second)
			continue
		}
		if cancelRetries >= 40 {
			bot.openLogger.Error("Max retries reached in BuyOrder. Exiting...")
			break
		}

		currentPrice := bids[0][0]
		if currentPrice > maxPrice*1.002 {
			bot.openLogger.WithFields(logrus.Fields{
				"current_price": currentPrice,
				"max_price":     maxPrice,
			}).Warn("Best buy price exceeds maximum limit, stopping.")
			_ = nobitex.CancelOrder(bot.apiToken, prevOrderID)
			break
		}

		if prevOrderID != 0 {
			// Price changed, or we want to adjust
			if len(bids) > 1 {
				nextBestBid := bids[1][0]
				// If our placed price is below next best or below current best
				if prevOrderPrice < nextBestBid*1.001 || prevOrderPrice < bids[0][0] {
					status, matched, err := nobitex.CheckOrderStatus(bot.apiToken, prevOrderID)
					if err != nil {
						bot.openLogger.WithFields(logrus.Fields{
							"order_id": prevOrderID,
						}).WithError(err).Error("Error checking buy order status")
						cancelRetries++
						time.Sleep(2 * time.Second)
						continue
					}
					if status == "Done" {
						bot.openLogger.WithField("order_id", prevOrderID).Info("Buy order fully matched. Exiting...")
						break
					}
					totalRemaining -= matched * prevPrice

					if err := nobitex.CancelOrder(bot.apiToken, prevOrderID); err != nil {
						bot.openLogger.WithField("order_id", prevOrderID).
							WithError(err).Error("Failed to cancel buy order")
						cancelRetries++
						time.Sleep(2 * time.Second)
						continue
					}
					bot.openLogger.WithField("order_id", prevOrderID).Info("Previous buy order canceled")
					prevOrderID = 0
				} else {
					continue
				}
			}
		}

		if totalRemaining <= 100000 {
			bot.openLogger.WithField("remaining_amount", totalRemaining/currentPrice).
				Info("Remaining funds too low. Stopping buy loop.")
			break
		}

		newPrice := bids[1][0] * 1.00001
		amount := totalRemaining / currentPrice
		orderID, err := nobitex.PlaceMarginOrder(bot.apiToken, bot.currencyPair, Leverage, "buy", amount, newPrice)
		if err != nil {
			bot.openLogger.WithError(err).WithFields(logrus.Fields{
				"retry":  cancelRetries,
				"amount": amount,
				"price":  newPrice,
			}).Error("Error placing buy order")
			cancelRetries++
			time.Sleep(1 * time.Second)
			continue
		}

		prevOrderID = orderID
		prevOrderPrice = newPrice
		prevPrice = currentPrice

		bot.openLogger.WithFields(logrus.Fields{
			"order_id": orderID,
			"price":    newPrice,
			"amount":   amount,
		}).Info("Buy order placed")

		time.Sleep(5 * time.Second)
	}
}

func (bot *TradingBot) PlaceSellOrder(balance float64, minPrice float64) {
	var (
		prevOrderID    int
		prevOrderPrice float64
		prevPrice      float64
		totalRemaining = balance
		cancelRetries  = 0
	)

	bot.openLogger.WithFields(logrus.Fields{
		"action":    "PlaceSellOrder_start",
		"balance":   balance,
		"min_price": minPrice,
	}).Info("Starting sell orders")

	for {
		bot.bookMutex.Lock()
		asks := bot.orderBookGlobal.Asks
		bot.bookMutex.Unlock()

		if len(asks) == 0 {
			bot.openLogger.Warn("No ask price available. Waiting...")
			time.Sleep(time.Second)
			continue
		}
		if cancelRetries >= 40 {
			bot.openLogger.Error("Max retries reached in SellOrder. Exiting...")
			break
		}

		currentPrice := asks[0][0]
		if currentPrice < minPrice*0.998 {
			bot.openLogger.WithFields(logrus.Fields{
				"current_price": currentPrice,
				"min_price":     minPrice,
			}).Warn("Best sell price is below the minimum limit, stopping.")
			_ = nobitex.CancelOrder(bot.apiToken, prevOrderID)
			break
		}

		if prevOrderID != 0 && len(asks) > 1 {
			nextBestAsk := asks[1][0]
			if prevOrderPrice > nextBestAsk*0.999 || prevOrderPrice > asks[0][0] {
				status, matched, err := nobitex.CheckOrderStatus(bot.apiToken, prevOrderID)
				if err != nil {
					bot.openLogger.WithFields(logrus.Fields{
						"order_id": prevOrderID,
					}).WithError(err).Error("Error checking sell order status")
					cancelRetries++
					time.Sleep(2 * time.Second)
					continue
				}
				if status == "Done" {
					bot.openLogger.WithField("order_id", prevOrderID).Info("Sell order fully matched. Exiting...")
					break
				}
				totalRemaining -= matched * prevPrice
				if totalRemaining <= 100000 {
					bot.openLogger.WithField("remaining_amount", totalRemaining).
						Info("Remaining funds too low. Stopping sell loop.")
					break
				}

				if err := nobitex.CancelOrder(bot.apiToken, prevOrderID); err != nil {
					bot.openLogger.WithField("order_id", prevOrderID).
						WithError(err).Error("Failed to cancel sell order")
					cancelRetries++
					time.Sleep(2 * time.Second)
					continue
				}
				bot.openLogger.WithField("order_id", prevOrderID).Info("Previous sell order canceled")
				prevOrderID = 0
			} else {
				continue
			}
		}

		if totalRemaining <= 100000 {
			bot.openLogger.WithField("remaining_amount", totalRemaining).
				Info("Remaining funds too low. Stopping sell loop.")
			break
		}

		newPrice := asks[1][0] * 0.99999
		amount := totalRemaining / currentPrice
		orderID, err := nobitex.PlaceMarginOrder(bot.apiToken, bot.currencyPair, Leverage, "sell", amount, newPrice)
		if err != nil {
			bot.openLogger.WithError(err).WithFields(logrus.Fields{
				"retry":  cancelRetries,
				"amount": amount,
				"price":  newPrice,
			}).Error("Error placing sell order")
			cancelRetries++
			time.Sleep(1 * time.Second)
			continue
		}

		prevOrderID = orderID
		prevOrderPrice = newPrice
		prevPrice = currentPrice

		bot.openLogger.WithFields(logrus.Fields{
			"order_id": orderID,
			"price":    newPrice,
			"amount":   amount,
		}).Info("Sell order placed")

		time.Sleep(5 * time.Second)
	}
}
