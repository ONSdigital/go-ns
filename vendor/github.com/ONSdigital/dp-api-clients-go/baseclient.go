package clients

import (
	"sync"
	"github.com/ONSdigital/dp-rchttp"
)

// APIClient represents a common structure for any api client
type APIClient struct {
	BaseURL    string
	HTTPClient rchttp.Clienter
	Lock       sync.RWMutex
}

