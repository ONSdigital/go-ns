package rchttp

import (
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// Client is an extension of the net/http client with ability to add
// timeouts, exponential backoff and context-based cancellation
type Client struct {
	MaxRetries         int
	ExponentialBackoff bool
	RetryTime          time.Duration
	HTTPClient         *http.Client
}

// DefaultClient is a go-ns specific http client with sensible timeouts,
// exponential backoff, and a contextual dialer
var DefaultClient = &Client{
	MaxRetries:         10,
	ExponentialBackoff: true,
	RetryTime:          20 * time.Millisecond,

	HTTPClient: &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
		},
	},
}

// Do calls ctxhttp.Do with the addition of exponential backoff
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	doer := func(args ...interface{}) (*http.Response, error) {
		return ctxhttp.Do(args[0].(context.Context), args[1].(*http.Client), args[2].(*http.Request))
	}

	resp, err := doer(ctx, c.HTTPClient, req)
	if err != nil {
		if c.ExponentialBackoff {
			return c.backoff(doer, ctx, c.HTTPClient, req)
		}
		return nil, err
	}

	if c.ExponentialBackoff && resp.StatusCode >= http.StatusInternalServerError {
		return c.backoff(doer, ctx, c.HTTPClient, req)
	}

	return resp, err
}

// Get calls Do with a GET
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req)
}

// Head calls Do with a HEAD
func (c *Client) Head(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(ctx, req)
}

// Post calls Do with a POST and the appropriate content-type and body
func (c *Client) Post(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	return c.Do(ctx, req)
}

// PostForm calls Post with the appropriate form content-type
func (c *Client) PostForm(ctx context.Context, uri string, data url.Values) (*http.Response, error) {
	return c.Post(ctx, uri, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *Client) backoff(f func(...interface{}) (*http.Response, error), args ...interface{}) (resp *http.Response, err error) {
	for attempt := 1; attempt <= c.MaxRetries; attempt++ {
		// ensure that the context is not cancelled before iterating
		if args[0].(context.Context).Err() != nil {
			err = args[0].(context.Context).Err()
			return
		}

		time.Sleep(getSleepTime(attempt, c.RetryTime))

		resp, err = f(args...)
		// prioritise any context cancellation
		if args[0].(context.Context).Err() != nil {
			err = args[0].(context.Context).Err()
			return
		}
		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			return
		}
	}
	return
}

// getSleepTime will return a sleep time based on the attempt and initial retry time.
// It uses the algorithm 2^n where n is the attempt number (double the previous) and
// a randomization factor of between 0-5ms so that the server isn't being hit constantly
// at the same time by many clients
func getSleepTime(attempt int, retryTime time.Duration) time.Duration {
	n := (math.Pow(2, float64(attempt)))
	rand.Seed(time.Now().Unix())
	rnd := time.Duration(rand.Intn(4)+1) * time.Millisecond
	return (time.Duration(n) * retryTime) - rnd
}
