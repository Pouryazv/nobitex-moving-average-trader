package nobitex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// performAuthenticatedRequest handles GET/POST requests with an auth token.
func performAuthenticatedRequest(apiToken, method, url string, payload interface{}) ([]byte, error) {
	var req *http.Request
	var err error

	if method == http.MethodGet {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	} else {
		jsonData, e := json.Marshal(payload)
		if e != nil {
			return nil, fmt.Errorf("payload marshal error: %v", e)
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	if err != nil {
		return nil, fmt.Errorf("request creation error: %v", err)
	}

	req.Header.Set("Authorization", "Token "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status: %d (%s) body=%s", resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
	}
	return body, nil
}
