package common

import "sync"

type APIClient struct {
	BaseURL    string
	AuthToken  string
	HTTPClient RCHTTPClienter
	Lock       sync.RWMutex
}
