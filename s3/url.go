package s3

import (
	"net/url"
	"strings"
)

// NewURL create a new instance of URL.
func NewURL(rawURL string) (*URL, error) {

	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return &URL{
		URL: url,
	}, nil
}

// URL represents a fully qualified S3 URL that includes the bucket name.
type URL struct {
	*url.URL
}

// BucketName returns the bucket name from the URL.
func (url *URL) BucketName() string {
	return url.Host
}

// Path returns the file path from the URL.
func (url *URL) Path() string {
	return strings.TrimPrefix(url.URL.Path, "/")
}
