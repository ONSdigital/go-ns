package rhttp

import (
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Client is an extension of the net/http client with ability to add
// timeouts and exponential backoff
type Client struct {
	MaxRetries         int
	ExponentialBackoff bool
	RetryTime          time.Duration

	HTTPClient *http.Client
}

// DefaultClient is a go-ns specific http client with sensible timeouts
// and exponential backoff
var DefaultClient = &Client{
	MaxRetries:         10,
	ExponentialBackoff: true,
	RetryTime:          20 * time.Millisecond,

	HTTPClient: &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
		},
	},
}

// Do calls net/http Do with the addition of exponential backoff
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	do := func(args ...interface{}) (*http.Response, error) {
		return c.HTTPClient.Do(args[0].(*http.Request))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if c.ExponentialBackoff {
			return c.backoff(do, req)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError && c.ExponentialBackoff {
		return c.backoff(do, req)
	}

	return resp, err
}

// Get calls net/http Get with the addition of exponential backoff
func (c *Client) Get(url string) (*http.Response, error) {
	get := func(args ...interface{}) (*http.Response, error) {
		return c.HTTPClient.Get(args[0].(string))
	}

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		if c.ExponentialBackoff {
			return c.backoff(get, url)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError && c.ExponentialBackoff {
		return c.backoff(get, url)
	}

	return resp, err
}

// Head calls net/http Head with the addition of exponential backoff
func (c *Client) Head(url string) (*http.Response, error) {
	head := func(args ...interface{}) (*http.Response, error) {
		return c.HTTPClient.Head(args[0].(string))
	}

	resp, err := c.HTTPClient.Head(url)
	if err != nil {
		if c.ExponentialBackoff {
			return c.backoff(head, url)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError && c.ExponentialBackoff {
		return c.backoff(head, url)
	}

	return resp, err
}

// Post calls net/http Post with the addition of exponential backoff
func (c *Client) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	post := func(args ...interface{}) (*http.Response, error) {
		return c.HTTPClient.Post(args[0].(string), args[1].(string), args[2].(io.Reader))
	}

	resp, err := c.HTTPClient.Post(url, contentType, body)
	if err != nil {
		if c.ExponentialBackoff {
			return c.backoff(post, url, contentType, body)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError && c.ExponentialBackoff {
		return c.backoff(post, url, contentType, body)
	}

	return resp, err
}

// PostForm calls net/http PostForm with the addition of exponential backoff
func (c *Client) PostForm(uri string, data url.Values) (*http.Response, error) {
	postForm := func(args ...interface{}) (*http.Response, error) {
		return c.HTTPClient.PostForm(args[0].(string), args[1].(url.Values))
	}

	resp, err := c.HTTPClient.PostForm(uri, data)
	if err != nil {
		if c.ExponentialBackoff {
			return c.backoff(postForm, uri, data)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError && c.ExponentialBackoff {
		return c.backoff(postForm, uri, data)
	}

	return resp, err
}

func (c *Client) backoff(f func(...interface{}) (*http.Response, error), args ...interface{}) (resp *http.Response, err error) {
	for attempt := 1; attempt <= c.MaxRetries; attempt++ {
		sleepTime := getSleepTime(attempt, c.RetryTime)

		time.Sleep(sleepTime)

		resp, err = f(args...)
		if err != nil || resp.StatusCode >= http.StatusInternalServerError {
			if attempt == c.MaxRetries {
				return
			}
			continue
		}

		return
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
