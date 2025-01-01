package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"nobitex-sma-bot/internal/nobitex"
	"strconv"
	"strings"
	"time"
)

// MonitorPositionsAndClose fetches open positions, logs them, and places OCO orders if needed.
func (bot *TradingBot) MonitorPositionsAndClose() {
	positions, err := nobitex.GetOpenPositions(bot.apiToken, strings.ToLower(strings.TrimSuffix(bot.currencyPair, "IRT")))
	if err != nil {
		bot.closeLogger.WithError(err).Error("Error fetching positions")
		return
	}
	bot.positionCount = len(positions)
	bot.closeLogger.WithField("count", bot.positionCount).Info("Open positions fetched")

	// Reset local counter for balance in positions
	bot.posMutex.Lock()
	bot.balanceInPositions = 0.0
	bot.posMutex.Unlock()

	for _, pos := range positions {
		positionID := pos.ID

		entryPrice, err := strconv.ParseFloat(pos.EntryPrice, 64)
		if err != nil {
			bot.closeLogger.WithFields(logrus.Fields{
				"position_id": positionID,
				"entry_price": pos.EntryPrice,
				"error":       err.Error(),
			}).Error("Error parsing entry price")
			continue
		}
		totalAsset, err := strconv.ParseFloat(pos.TotalAsset, 64)
		if err == nil {
			bot.posMutex.Lock()
			bot.balanceInPositions += totalAsset
			bot.posMutex.Unlock()
		}

		bot.ocoMu.Lock()
		ocoExists := bot.ocoOrders[positionID]
		bot.ocoMu.Unlock()

		bot.priceMu.RLock()
		bestBid := bot.bidBest
		bestAsk := bot.askBest
		bot.priceMu.RUnlock()

		if bestBid == 0 || bestAsk == 0 {
			continue
		}

		// If no OCO order placed yet for this position
		if !ocoExists {
			var takeProfitPrice, stopLossPrice float64
			switch pos.Side {
			case "buy":
				takeProfitPrice = max(entryPrice*(1+ProfitTarget), bestBid)
				stopLossPrice = min(entryPrice*(1-StopLoss), bestBid)
			case "sell":
				takeProfitPrice = min(entryPrice*(1-ProfitTarget), bestAsk)
				stopLossPrice = max(entryPrice*(1+StopLoss), bestAsk)
			}

			liability, err := strconv.ParseFloat(pos.Liability, 64)
			if err != nil {
				bot.closeLogger.WithError(err).WithField("position_id", positionID).
					Error("Error parsing liability for OCO")
				continue
			}

			orderID, err := bot.ClosePositionOrder(positionID, liability, takeProfitPrice, stopLossPrice)
			if err != nil {
				bot.closeLogger.WithFields(logrus.Fields{
					"position_id": positionID,
					"error":       err.Error(),
				}).Error("Failed to place OCO")
				continue
			}

			bot.ocoMu.Lock()
			bot.ocoOrders[positionID] = true
			bot.ocoMu.Unlock()

			bot.closeLogger.WithFields(logrus.Fields{
				"position_id": positionID,
				"order_id":    orderID,
				"take_profit": takeProfitPrice,
				"stop_loss":   stopLossPrice,
			}).Info("OCO placed for position")
		}

		// Schedule a check to see if the position got closed
		go func(pos nobitex.Position) {
			time.Sleep(time.Minute)
			closed, err := bot.IsPositionClosed(pos.ID)
			if err != nil {
				bot.closeLogger.WithFields(logrus.Fields{
					"position_id": pos.ID,
					"error":       err.Error(),
				}).Error("Error checking position status")
				return
			}
			if closed {
				bot.ocoMu.Lock()
				delete(bot.ocoOrders, pos.ID)
				bot.ocoMu.Unlock()
				bot.closeLogger.WithField("position_id", pos.ID).Info("Position closed. Removed from OCO.")
			}
		}(pos)
	}
}

// ClosePositionOrder places an OCO order to close a position, with retries.
func (bot *TradingBot) ClosePositionOrder(
	positionID int,
	amount, takeProfitPrice, stopLossPrice float64,
) (int, error) {
	url := fmt.Sprintf("https://api.nobitex.ir/positions/%d/close", positionID)
	adjustment := 0.9
	if takeProfitPrice < stopLossPrice {
		adjustment = 1.1
	}
	payload := map[string]interface{}{
		"amount":         strconv.FormatFloat(amount, 'f', -1, 64),
		"price":          strconv.FormatFloat(takeProfitPrice, 'f', -1, 64),
		"mode":           "oco",
		"stopPrice":      strconv.FormatFloat(stopLossPrice, 'f', -1, 64),
		"stopLimitPrice": strconv.FormatFloat(stopLossPrice*adjustment, 'f', -1, 64),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		bot.closeLogger.WithError(err).
			WithField("position_id", positionID).
			Error("Error marshaling OCO payload")
		return 0, err
	}

	const (
		maxRetries = 15
		retryDelay = 2 * time.Second
	)
	var orderID int
	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
		if err != nil {
			bot.closeLogger.WithFields(logrus.Fields{"position_id": positionID, "attempt": attempt}).
				WithError(err).Error("Error creating request to close position")
		} else {
			req.Header.Set("Authorization", "Token "+bot.apiToken)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				bot.closeLogger.WithFields(logrus.Fields{"position_id": positionID, "attempt": attempt}).
					WithError(err).Error("Error sending HTTP request for ClosePositionOrder")
			} else {
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					bot.closeLogger.WithFields(logrus.Fields{"position_id": positionID, "attempt": attempt}).
						WithError(err).Error("Error reading response for ClosePositionOrder")
				} else {
					var r struct {
						Status string `json:"status"`
						Order  struct {
							ID int `json:"id"`
						} `json:"order"`
						Orders []struct {
							ID int `json:"id"`
						} `json:"orders"`
					}
					if err := json.Unmarshal(body, &r); err != nil {
						bot.closeLogger.WithFields(logrus.Fields{"position_id": positionID, "attempt": attempt}).
							WithError(err).Error("Error unmarshaling response for ClosePositionOrder")
					} else if r.Status != "ok" {
						bot.closeLogger.WithFields(logrus.Fields{
							"position_id": positionID,
							"attempt":     attempt,
							"response":    string(body),
						}).Error("Close position OCO order failed (status != ok)")
					} else {
						if len(r.Orders) > 0 {
							orderID = r.Orders[0].ID
						} else {
							orderID = r.Order.ID
						}
						bot.closeLogger.WithFields(logrus.Fields{
							"position_id": positionID,
							"order_id":    orderID,
							"take_profit": takeProfitPrice,
							"stop_loss":   stopLossPrice,
							"attempt":     attempt,
						}).Info("OCO order placed successfully")
						return orderID, nil
					}
				}
			}
		}

		if attempt < maxRetries {
			bot.closeLogger.WithFields(logrus.Fields{
				"position_id": positionID,
				"attempt":     attempt,
			}).Warnf("Retrying OCO order in %v...", retryDelay)
			time.Sleep(retryDelay)
		} else {
			bot.closeLogger.WithFields(logrus.Fields{
				"position_id": positionID,
			}).Error("Max retries reached for OCO order")
		}
	}
	return 0, fmt.Errorf("failed to place OCO order after %d attempts", maxRetries)
}

// IsPositionClosed returns whether a position has status "Closed".
func (bot *TradingBot) IsPositionClosed(positionID int) (bool, error) {
	pos, err := nobitex.GetPositionDetails(bot.apiToken, positionID)
	if err != nil {
		return false, err
	}
	return pos.Status == "Closed", nil
}
