package clobclient

import "fmt"

// PolyException is the base error type.
type PolyException struct {
	Msg string
}

func (e *PolyException) Error() string {
	return e.Msg
}

// PolyAPIException represents an API error response.
type PolyAPIException struct {
	StatusCode int
	ErrorMsg   string
}

func (e *PolyAPIException) Error() string {
	return fmt.Sprintf("PolyApiException[status_code=%d, error_message=%s]", e.StatusCode, e.ErrorMsg)
}
