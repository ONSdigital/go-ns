package rhttp

import (
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ONSdigital/go-ns/log"
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
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if c.ExponentialBackoff {
			return c.doWithBackoff(req)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return c.doWithBackoff(req)
	}

	return resp, err
}

func (c *Client) doWithBackoff(req *http.Request) (resp *http.Response, err error) {
	for i := 0; i < c.MaxRetries; i++ {
		attempt := i + 1
		sleepTime := getSleepTime(attempt, c.RetryTime)

		time.Sleep(sleepTime)

		resp, err = c.HTTPClient.Do(req)
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

// Get calls net/http Get with the addition of exponential backoff
func (c *Client) Get(url string) (*http.Response, error) {
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		if c.ExponentialBackoff {
			return c.getWithBackoff(url)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return c.getWithBackoff(url)
	}

	return resp, err
}

func (c *Client) getWithBackoff(url string) (resp *http.Response, err error) {
	for i := 0; i < c.MaxRetries; i++ {
		attempt := i + 1
		sleepTime := getSleepTime(attempt, c.RetryTime)

		log.Debug("retrying....", log.Data{"attempt": attempt, "sleep time": sleepTime})

		time.Sleep(sleepTime)

		resp, err = c.HTTPClient.Get(url)
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

// Head calls net/http Head with the addition of exponential backoff
func (c *Client) Head(url string) (*http.Response, error) {
	resp, err := c.HTTPClient.Head(url)
	if err != nil {
		if c.ExponentialBackoff {
			return c.headWithBackoff(url)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return c.headWithBackoff(url)
	}

	return resp, err
}

func (c *Client) headWithBackoff(url string) (resp *http.Response, err error) {
	for i := 0; i < c.MaxRetries; i++ {
		attempt := i + 1
		sleepTime := getSleepTime(attempt, c.RetryTime)

		time.Sleep(sleepTime)

		resp, err = c.HTTPClient.Head(url)
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

// Post calls net/http Post with the addition of exponential backoff
func (c *Client) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	resp, err := c.HTTPClient.Post(url, contentType, body)
	if err != nil {
		if c.ExponentialBackoff {
			return c.postWithBackoff(url, contentType, body)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return c.postWithBackoff(url, contentType, body)
	}

	return resp, err
}

func (c *Client) postWithBackoff(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	for i := 0; i < c.MaxRetries; i++ {
		attempt := i + 1
		sleepTime := getSleepTime(attempt, c.RetryTime)

		time.Sleep(sleepTime)

		resp, err = c.HTTPClient.Post(url, contentType, body)
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

// PostForm calls net/http PostForm with the addition of exponential backoff
func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	resp, err := c.HTTPClient.PostForm(url, data)
	if err != nil {
		if c.ExponentialBackoff {
			return c.postFormWithBackoff(url, data)
		}
		return nil, err
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return c.postFormWithBackoff(url, data)
	}

	return resp, err
}

func (c *Client) postFormWithBackoff(url string, data url.Values) (resp *http.Response, err error) {
	for i := 0; i < c.MaxRetries; i++ {
		attempt := i + 1
		sleepTime := getSleepTime(attempt, c.RetryTime)

		time.Sleep(sleepTime)

		resp, err = c.HTTPClient.PostForm(url, data)
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

func getSleepTime(attempt int, retryTime time.Duration) time.Duration {
	n := (math.Pow(2, float64(attempt)))
	rand.Seed(time.Now().Unix())
	rnd := time.Duration(rand.Intn(4)+1) * time.Millisecond
	return (time.Duration(n) * retryTime) - rnd
}
