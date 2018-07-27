package rchttp

import (
	"errors"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ONSdigital/go-ns/common"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

// Client is an extension of the net/http client with ability to add
// timeouts, exponential backoff and context-based cancellation
type Client struct {
	MaxRetries           int
	ExponentialBackoff   bool
	RetryTime            time.Duration
	HTTPClient           *http.Client
	AuthToken            string
	DownloadServiceToken string
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

// NewClient returns a copy of DefaultClient
func NewClient() common.RCHTTPClienter {
	newClient := *DefaultClient
	return &newClient
}

// ClientWithTimeout facilitates creating a client and setting request timeout
func ClientWithTimeout(c common.RCHTTPClienter, timeout time.Duration) common.RCHTTPClienter {
	if c == nil {
		c = NewClient()
	}
	c.SetTimeout(timeout)
	return c
} // ClientWithTimeout facilitates creating a client and setting request timeout

func (c *Client) SetTimeout(timeout time.Duration) {
	c.HTTPClient.Timeout = timeout
}

// ClientWithServiceToken facilitates creating a client and setting service auth
func ClientWithServiceToken(c common.RCHTTPClienter, authToken string) common.RCHTTPClienter {
	if c == nil {
		c = NewClient()
	}
	c.SetAuthToken(authToken)
	return c
}
func (c *Client) SetAuthToken(authToken string) {
	c.AuthToken = authToken
}

// ClientWithDownloadServiceToken facilitates creating a client and setting service auth
func ClientWithDownloadServiceToken(c common.RCHTTPClienter, token string) common.RCHTTPClienter {
	if c == nil {
		c = NewClient()
	}
	c.SetDownloadServiceToken(token)
	return c
}
func (c *Client) SetDownloadServiceToken(token string) {
	c.DownloadServiceToken = token
}

func (c *Client) GetMaxRetries() int {
	return c.MaxRetries
}
func (c *Client) SetMaxRetries(maxRetries int) {
	c.MaxRetries = maxRetries
}

// Do calls ctxhttp.Do with the addition of exponential backoff
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {

	if len(c.AuthToken) > 0 {
		// only add this header if not already set (e.g. for authClient)
		if len(req.Header.Get(common.AuthHeaderKey)) == 0 {
			common.AddServiceTokenHeader(req, c.AuthToken)
		}
	}
	if len(c.DownloadServiceToken) > 0 {
		// only add this header if not already set
		if len(req.Header.Get(common.DownloadServiceHeaderKey)) == 0 {
			common.AddDownloadServiceTokenHeader(req, c.DownloadServiceToken)
		}
	}
	if common.IsUserPresent(ctx) {
		// only add this header if not already set
		if len(req.Header.Get(common.UserHeaderKey)) == 0 {
			common.AddUserHeader(req, common.User(ctx))
		}
	}
	if common.IsFlorenceIdentityPresent(ctx) {
		common.SetFlorenceHeader(ctx, req)
	}

	// get any existing correlation-id (might be "id1,id2"), append a new one, add to headers
	upstreamCorrelationIds := common.GetRequestId(ctx)
	addedIdLen := 20
	if upstreamCorrelationIds != "" {
		// get length of (first of) IDs (e.g. "id1" is 3), new ID will be half that size
		addedIdLen = len(upstreamCorrelationIds) / 2
		if commaPosition := strings.Index(upstreamCorrelationIds, ","); commaPosition > 1 {
			addedIdLen = commaPosition / 2
		}
		upstreamCorrelationIds += ","
	}
	common.AddRequestIdHeader(req, upstreamCorrelationIds+common.NewRequestID(addedIdLen))

	doer := func(args ...interface{}) (*http.Response, error) {
		req := args[2].(*http.Request)
		if req.ContentLength > 0 {
			var err error
			req.Body, err = req.GetBody()
			if err != nil {
				return nil, err
			}
		}
		return ctxhttp.Do(args[0].(context.Context), args[1].(*http.Client), req)
	}

	resp, err := doer(ctx, c.HTTPClient, req)
	if err != nil {
		if c.ExponentialBackoff {
			return c.backoff(doer, err, ctx, c.HTTPClient, req)
		}
		return nil, err
	}

	if c.ExponentialBackoff {
		if resp.StatusCode >= http.StatusInternalServerError {
			return c.backoff(doer, err, ctx, c.HTTPClient, req, errors.New("Bad server status"))
		}

		if resp.StatusCode == http.StatusConflict {
			return c.backoff(doer, err, ctx, c.HTTPClient, req, errors.New("Conflict - request could not be completed due to a conflict with the current state of the target resource"))
		}
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

// Put calls Do with a PUT and the appropriate content-type and body
func (c *Client) Put(ctx context.Context, url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("PUT", url, body)
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

func (c *Client) backoff(f func(...interface{}) (*http.Response, error), retryErr error, args ...interface{}) (resp *http.Response, err error) {
	if c.GetMaxRetries() < 1 {
		return nil, retryErr
	}
	for attempt := 1; attempt <= c.GetMaxRetries(); attempt++ {
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
		if err == nil && resp.StatusCode < http.StatusInternalServerError && resp.StatusCode != http.StatusConflict {
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
