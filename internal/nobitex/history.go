package nobitex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type OHLCVHistory struct {
	Status string    `json:"s"`
	Close  []float64 `json:"c"`
}

func GetOHLCVData(symbol, resolution string, from, to int64) ([]float64, error) {
	url := fmt.Sprintf(
		"https://api.nobitex.ir/market/udf/history?symbol=%s&resolution=%s&from=%d&to=%d",
		symbol, resolution, from, to,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OHLCV data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OHLCV response: %w", err)
	}

	var data OHLCVHistory
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse OHLCV JSON: %w", err)
	}
	if data.Status != "ok" {
		return nil, fmt.Errorf("Nobitex OHLCV error: status=%s", data.Status)
	}

	return data.Close, nil
}
