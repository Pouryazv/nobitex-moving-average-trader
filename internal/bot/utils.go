package bot

import (
	"nobitex-sma-bot/internal/nobitex"
	"strconv"
	"time"
)

func (bot *TradingBot) calculateSMA(prices []float64) float64 {
	var sum float64
	for _, p := range prices {
		sum += p
	}
	return sum / float64(len(prices))
}

func (bot *TradingBot) parseOrderBook(raw [][]string) [][2]float64 {
	var result [][2]float64
	for _, entry := range raw {
		if len(entry) < 2 {
			continue
		}
		price, err1 := strconv.ParseFloat(entry[0], 64)
		amount, err2 := strconv.ParseFloat(entry[1], 64)
		if err1 == nil && err2 == nil {
			result = append(result, [2]float64{price, amount})
		}
	}
	return result
}

func (b *TradingBot) fetchOHLCVData() ([]float64, error) {
	endTime := time.Now().Unix()
	startTime := endTime - 60*Pastmin // e.g., fetch last 20 minutes of data

	return nobitex.GetOHLCVData(
		b.currencyPair,
		"1", // 1-minute resolution, or whichever you want
		startTime,
		endTime,
	)
}
