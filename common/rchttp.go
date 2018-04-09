package common

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

//go:generate mockgen -destination mock_common/rchttp_client.go github.com/ONSdigital/go-ns/common RCHTTPClient
//go:generate moq -out commontest/rchttp_client.go -pkg commontest . RCHTTPClient

// RCHTTPClient provides an interface for methods on an HTTP Client
type RCHTTPClienter interface {
	SetAuthToken(authToken string)
	SetTimeout(timeout time.Duration)
	SetMaxRetries(int)
	GetMaxRetries() int

	Get(ctx context.Context, url string) (*http.Response, error)
	Head(ctx context.Context, url string) (*http.Response, error)
	Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error)
	PostForm(ctx context.Context, uri string, data url.Values) (*http.Response, error)

	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}
