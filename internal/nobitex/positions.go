package nobitex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetOpenPositions retrieves all active positions for the given srcCurrency.
func GetOpenPositions(apiToken, srcCurrency string) ([]Position, error) {
	url := fmt.Sprintf("https://api.nobitex.ir/positions/list?srcCurrency=%s&status=active", srcCurrency)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response PositionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if response.Status != "ok" {
		return nil, fmt.Errorf("failed to fetch positions, got: %v", response)
	}
	return response.Positions, nil
}

// GetPositionDetails fetches details of a specific position by ID.
func GetPositionDetails(apiToken string, positionID int) (*Position, error) {
	url := fmt.Sprintf("https://api.nobitex.ir/positions/%d/status", positionID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Status   string   `json:"status"`
		Position Position `json:"position"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	if data.Status != "ok" {
		return nil, fmt.Errorf("failed to fetch position details, status: %s", data.Status)
	}
	return &data.Position, nil
}
