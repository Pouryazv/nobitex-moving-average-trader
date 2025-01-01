package nobitex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// CancelOrder attempts to cancel an existing order by its ID.
func CancelOrder(apiToken string, orderID int) error {
	url := baseURL + updateOrderStatusEndpoint
	payload := map[string]interface{}{
		"order":  orderID,
		"status": "canceled",
	}
	respData, err := performAuthenticatedRequest(apiToken, http.MethodPost, url, payload)
	if err != nil {
		return err
	}

	var cancelResp CancelOrderResponse
	if err := json.Unmarshal(respData, &cancelResp); err != nil {
		return err
	}
	if cancelResp.Status != "ok" {
		return fmt.Errorf("failed to cancel: code=%s, msg=%s", cancelResp.Code, cancelResp.Message)
	}

	return nil
}

// PlaceMarginOrder creates a margin limit order (buy or sell) and returns its ID.
func PlaceMarginOrder(apiToken, currencyPair, leverage, orderType string, amount, price float64) (int, error) {
	url := baseURL + placeMarginOrderEndpoint

	src, dst, err := splitCurrencyPair(currencyPair)
	if err != nil {
		return 0, err
	}
	if orderType != "buy" && orderType != "sell" {
		return 0, fmt.Errorf("invalid order type: %s", orderType)
	}

	payload := map[string]interface{}{
		"execution":   "limit",
		"srcCurrency": src,
		"dstCurrency": dst,
		"type":        orderType,
		"leverage":    leverage,
		"amount":      fmt.Sprintf("%.8f", amount),
		"price":       fmt.Sprintf("%.0f", price),
	}
	respData, err := performAuthenticatedRequest(apiToken, http.MethodPost, url, payload)
	if err != nil {
		return 0, err
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(respData, &orderResp); err != nil {
		return 0, err
	}
	if orderResp.Status != "ok" {
		return 0, fmt.Errorf("failed to place %s order: code=%s, msg=%s", orderType, orderResp.Code, orderResp.Message)
	}

	return orderResp.Order.ID, nil
}

func CheckOrderStatus(apiToken string, orderID int) (string, float64, error) {
	url := "https://api.nobitex.ir/market/orders/status"
	payload := map[string]interface{}{
		"id": orderID,
	}

	responseData, err := performAuthenticatedRequest(apiToken, http.MethodPost, url, payload)
	if err != nil {
		return "", 0, err
	}

	var statusResponse struct {
		Status string `json:"status"`
		Order  struct {
			Status        string `json:"status"`
			MatchedAmount string `json:"matchedAmount"`
			Unmatched     string `json:"unmatchedAmount"`
		} `json:"order"`
	}

	if err := json.Unmarshal(responseData, &statusResponse); err != nil {
		return "", 0, err
	}

	// Convert matched amount
	matchedAmount, _ := strconv.ParseFloat(statusResponse.Order.MatchedAmount, 64)
	return statusResponse.Order.Status, matchedAmount, nil
}
func splitCurrencyPair(pair string) (string, string, error) {
	dstCurrencies := []string{"IRT", "USDT", "BTC", "ETH", "USDC", "BNB", "DOGE"}
	for _, dst := range dstCurrencies {
		if strings.HasSuffix(pair, dst) {
			src := strings.TrimSuffix(pair, dst)
			if src == "" {
				return "", "", fmt.Errorf("invalid pair: %s", pair)
			}
			return strings.ToLower(src), "rls", nil
		}
	}
	return "", "", fmt.Errorf("couldn't split pair: %s", pair)
}
