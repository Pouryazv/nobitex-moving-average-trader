package main

import (
	"log"
	"nobitex-sma-bot/internal/bot"
	"os"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <CurrencyPair>")
	}
	currencyPair := os.Args[1]

	tradingBot := bot.NewTradingBot(currencyPair)
	tradingBot.Run()
}
