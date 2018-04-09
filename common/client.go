package common

import "sync"

type APIClient struct {
	BaseURL    string
	AuthToken  string
	HTTPClient RCHTTPClienter
	Lock       sync.RWMutex
}

type APIClienter interface {
	NewAPIClient(cli RCHTTPClienter, url, authToken string)
}
