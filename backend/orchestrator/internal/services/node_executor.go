package services

import (
	"net/http"
)

// Rename struct!
type DefaultNodeExecutor struct {
	httpClient *http.Client
}

// Optionally add constructor
func NewDefaultNodeExecutor() *DefaultNodeExecutor {
	return &DefaultNodeExecutor{
		httpClient: &http.Client{},
	}
}
