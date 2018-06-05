package common

import "sync"

// APIClient represents a common structure for any api client
type APIClient struct {
	BaseURL    string
	AuthToken  string
	HTTPClient RCHTTPClienter
	Lock       sync.RWMutex
}
