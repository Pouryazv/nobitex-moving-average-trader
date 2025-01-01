# 📈 Nobitex SMA Trading Bot  
**Automated Trading Bot for Nobitex using Simple Moving Average (SMA) Strategy**  

## 🛠️ Overview  
This project is a trading bot designed to automate buying and selling on **Nobitex** (cryptocurrency exchange) based on the **Simple Moving Average (SMA)** strategy. The bot monitors price deviations and executes buy/sell orders when conditions are met, helping traders automate their strategies.
Due to the high volatility and rapid price fluctuations on Nobitex, manual trading can lead to missed opportunities or late reactions. This bot automates the trading process, ensuring faster responses to price deviations and better risk management. By using the SMA strategy, the bot can take advantage of market trends and price corrections, providing a structured and disciplined approach to trading even in unpredictable market conditions.

## 🚀 Features  
- **Automated Buy/Sell Orders** – Trades are executed based on deviations from the SMA.  
- **WebSocket Integration** – Real-time order book updates for fast and accurate trading.  
- **OCO Orders** – One-Cancels-the-Other (OCO) orders to manage risk.  
- **Position Monitoring** – Automatically closes open positions when the target profit or stop-loss is triggered.  
- **Logging** – Detailed logging of all trades and operations for transparency.  
- **Concurrency Handling** – Efficient handling of simultaneous buy/sell operations.
  

## 🛠️ Installation and Setup  

### 1. Clone the Repository  
```bash
git clone https://github.com/username/nobitex-sma-bot.git
cd nobitex-sma-bot
```

### 2. Set Up Environment Variables  
Create a `.env` file in the root directory and add your Nobitex API token:  

```plaintext
NOBITEX_API_TOKEN=your_api_token_here
```

### 3. Install Dependencies  
```bash
go mod tidy
```

### 4. Build the Project  
```bash
go build ./cmd
```

### 5. Run the Bot  
```bash
go run ./cmd/main.go BTCIRT
```
- Replace `BTCIRT` with the trading pair of your choice (e.g., `ETHUSDT`, `DOGEIRT`).


## ⚙️ Configuration  
- **Leverage:** Set the leverage value in the code (default is `3.0`).  
- **Price Deviation:** Adjust the price deviation to control sensitivity for trades.  
- **Profit/Stop-Loss:** Configure profit targets and stop-loss limits.  
- **Minimum Balance:** Minimum balance to trigger trades is set to `50,000,000 Rials`. This can be modified in the configuration.
- **Fetch Interval:** By default, the bot fetches data from the last 30 minutes. This interval can be adjusted by modifying the `Pastmin` variable in the config file.


## 📊 How It Works  
1. **Real-time Order Book Monitoring** – The bot subscribes to Nobitex’s WebSocket order book and continuously monitors price changes.  
2. **SMA Calculation** – The bot calculates the SMA based on the latest price data.  
3. **Trade Execution** – If the price deviates by more than the configured threshold from the SMA, the bot places buy or sell orders.  
4. **Risk Management** – The bot automatically places OCO orders to secure profits and limit losses.  
5. **Position Monitoring** – Open positions are tracked and closed if profit or stop-loss conditions are met.  


## ⚠️ Disclaimer  
This bot is intended for **educational and experimental purposes only**. Automated trading carries significant risks, and you may lose funds. Always test with small amounts or in a sandbox environment. The author is not responsible for any financial losses resulting from the use of this software.  


## 🐛 Troubleshooting  

- **WebSocket Disconnection:**  
   The bot will attempt to reconnect automatically. If persistent disconnections occur, check for Nobitex WebSocket downtime.  


## 🔧 Future Improvements  
- Support for multiple trading pairs simultaneously.  
- Dynamic leverage adjustment based on market conditions.  
- Machine learning-based decision-making for improved trading signals.  
- UI for monitoring trades in real-time.  


## 👨‍💻 Contributing  
Pull requests are welcome. For major changes, please open an issue to discuss the proposed changes.  

1. Fork the project.  
2. Create your feature branch (`git checkout -b feature/your-feature`).  
3. Commit your changes (`git commit -m 'Add new feature'`).  
4. Push to the branch (`git push origin feature/your-feature`).  
5. Open a Pull Request.  


## 📝 License  
This project is licensed under the **Apache License 2.0**.  
