package nobitex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// GetAvailableBalance returns the available RLS balance for margin trading.
func GetAvailableBalance(apiToken string) (float64, error) {
	url := baseURL + walletsEndpoint
	respData, err := performAuthenticatedRequest(apiToken, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	var balanceResp BalanceResponse
	if err := json.Unmarshal(respData, &balanceResp); err != nil {
		return 0, err
	}
	if balanceResp.Status != "ok" {
		return 0, fmt.Errorf("failed to get balance: code=%s, msg=%s", balanceResp.Code, balanceResp.Message)
	}

	wallet, found := balanceResp.Wallets["RLS"]
	if !found {
		return 0, fmt.Errorf("RLS wallet not found")
	}

	balance, err := strconv.ParseFloat(wallet.Balance, 64)
	if err != nil {
		return 0, fmt.Errorf("parse balance error: %v", err)
	}
	blocked, err := strconv.ParseFloat(wallet.Blocked, 64)
	if err != nil {
		return 0, fmt.Errorf("parse blocked error: %v", err)
	}

	return balance - blocked, nil
}
