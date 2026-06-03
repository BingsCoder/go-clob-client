package clobclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func overloadHeaders(req *http.Request, headers map[string]string) {
	req.Header.Set("User-Agent", "go_clob_client")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "application/json")
	if req.Method == http.MethodGet {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// httpRequest performs an HTTP request and returns the parsed JSON response.
func httpRequest(endpoint, method string, headers map[string]string, data interface{}, params map[string]string) (interface{}, error) {
	var bodyReader io.Reader
	if data != nil {
		switch v := data.(type) {
		case string:
			bodyReader = bytes.NewBufferString(v)
		case []byte:
			bodyReader = bytes.NewBuffer(v)
		default:
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewBuffer(jsonBytes)
		}
	}

	req, err := http.NewRequest(method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query params
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	overloadHeaders(req, headers)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, &PolyAPIException{ErrorMsg: fmt.Sprintf("Request exception: %s", err.Error())}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &PolyAPIException{ErrorMsg: "Failed to read response body"}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &PolyAPIException{
			StatusCode: resp.StatusCode,
			ErrorMsg:   string(respBody),
		}
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// Return raw text if not JSON
		return string(respBody), nil
	}
	return result, nil
}

func httpGet(endpoint string, headers map[string]string, params map[string]string) (interface{}, error) {
	return httpRequest(endpoint, http.MethodGet, headers, nil, params)
}

func httpPost(endpoint string, headers map[string]string, data interface{}, params map[string]string, retryOnError bool) (interface{}, error) {
	result, err := httpRequest(endpoint, http.MethodPost, headers, data, params)
	if err != nil && retryOnError {
		if isTransientError(err) {
			time.Sleep(30 * time.Millisecond)
			return httpRequest(endpoint, http.MethodPost, headers, data, params)
		}
	}
	return result, err
}

func httpDelete(endpoint string, headers map[string]string, data interface{}, params map[string]string) (interface{}, error) {
	return httpRequest(endpoint, http.MethodDelete, headers, data, params)
}

func httpPut(endpoint string, headers map[string]string, data interface{}, params map[string]string) (interface{}, error) {
	return httpRequest(endpoint, http.MethodPut, headers, data, params)
}

func isTransientError(err error) bool {
	if apiErr, ok := err.(*PolyAPIException); ok {
		if apiErr.StatusCode >= 500 && apiErr.StatusCode < 600 {
			return true
		}
		if apiErr.StatusCode == 0 {
			return true
		}
	}
	return false
}
