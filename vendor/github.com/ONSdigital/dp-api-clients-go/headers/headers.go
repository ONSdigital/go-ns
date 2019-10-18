// headers package provides header name constants and get/set functions for commonly used http headers in the
// dp-api-clients-go repo. Package replaces go-ns lib and should be treated as the single source of truth
package headers

import (
	"errors"
	"net/http"
	"strings"
)

const (
	// collectionIDHeader is the name used for a collection ID http request header
	collectionIDHeader = "Collection-Id"

	// userAuthTokenHeader is the user Florence auth token header name
	userAuthTokenHeader = "X-Florence-Token"

	// serviceAuthToken the service auth token header name
	serviceAuthTokenHeader = "Authorization"

	// bearerPrefix is the prefix for authorization header values
	bearerPrefix = "Bearer "

	// downloadServiceToken is the authorization header for the download service
	downloadServiceTokenHeader = "X-Download-Service-Token"

	// userIdentity is the user identity header used to forward a confirmed identity to another API.
	userIdentityHeader = "User-Identity"

	// requestIDHeader is the unique request ID header name.
	requestIDHeader = "X-Request-Id"

	// localeCodeHeader is the locale code header name
	localeCodeHeader = "LocaleCode"
)

var (
	// ErrHeaderNotFound returned if the requested header is not present in the provided request
	ErrHeaderNotFound = errors.New("header not found")

	// ErrValueEmpty returned if an empty value is passed to a SetX header function
	ErrValueEmpty = errors.New("header not set as value was empty")

	// ErrRequestNil return if SetX header function is called with a nil request
	ErrRequestNil = errors.New("error setting request header request was nil")
)

// IsErrNotFound return true if the err equal to ErrHeaderNotFound. Return false otherwise
func IsErrNotFound(err error) bool {
	return err == ErrHeaderNotFound
}

// IsNotErrNotFound return true if error is not nil and not equal to ErrHeaderNotFound
func IsNotErrNotFound(err error) bool {
	return !IsErrNotFound(err)
}

// GetCollectionID returns the value of the "Collection-Id" request header if it exists, returns ErrHeaderNotFound if
// the header is not found.
func GetCollectionID(req *http.Request) (string, error) {
	return getRequestHeader(req, collectionIDHeader)
}

// GetUserAuthToken returns the value of the "X-Florence-Token" request header if it exists, returns ErrHeaderNotFound if
// the header is not found.
func GetUserAuthToken(req *http.Request) (string, error) {
	return getRequestHeader(req, userAuthTokenHeader)
}

// GetServiceAuthToken returns the value of the "Authorization" request header if it exists, returns ErrHeaderNotFound if
// the header is not found. If the header exists the "Bearer " prefixed is removed from returned value.
func GetServiceAuthToken(req *http.Request) (string, error) {
	token, err := getRequestHeader(req, serviceAuthTokenHeader)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(token, bearerPrefix) {
		token = strings.TrimPrefix(token, bearerPrefix)
	}
	return token, nil
}

// GetDownloadServiceToken returns the value of the "X-Download-Service-Token" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetDownloadServiceToken(req *http.Request) (string, error) {
	return getRequestHeader(req, downloadServiceTokenHeader)
}

// GetUserIdentity returns the value of the "User-Identity" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetUserIdentity(req *http.Request) (string, error) {
	return getRequestHeader(req, userIdentityHeader)
}

// GetRequestID returns the value of the "X-Request-Id" request header if it exists, returns
// ErrHeaderNotFound if the header is not found.
func GetRequestID(req *http.Request) (string, error) {
	return getRequestHeader(req, requestIDHeader)
}

// GetLocaleCode returns the value of the "LocaleCode" request header if it exists, returns ErrHeaderNotFound if
// the header is not found.
func GetLocaleCode(req *http.Request) (string, error) {
	return getRequestHeader(req, localeCodeHeader)
}

func getRequestHeader(req *http.Request, headerName string) (string, error) {
	if req == nil {
		return "", ErrRequestNil
	}

	headerValue := req.Header.Get(headerName)
	if len(headerValue) == 0 {
		return "", ErrHeaderNotFound
	}

	return headerValue, nil
}

// SetCollectionID set the collection ID header on the provided request. If the collection ID header is already present
// in the request it will be overwritten by the new value. If the header value is empty returns ErrValueEmpty
func SetCollectionID(req *http.Request, headerValue string) error {
	return setRequestHeader(req, collectionIDHeader, headerValue)
}

// SetUserAuthToken set the user authentication token header on the provided request. If the authentication token is
// already present it will be overwritten by the new value. If the header value is empty returns ErrValueEmpty
func SetUserAuthToken(req *http.Request, headerValue string) error {
	return setRequestHeader(req, userAuthTokenHeader, headerValue)
}

// SetServiceAuthToken set the service authentication token header on the provided request. If the authentication token is
// already present it will be overwritten by the new value. If the header value is empty then returns ErrValueEmpty
func SetServiceAuthToken(req *http.Request, headerValue string) error {
	if req == nil {
		return ErrRequestNil
	}

	if len(headerValue) == 0 {
		return ErrValueEmpty
	}

	if !strings.HasPrefix(headerValue, bearerPrefix) {
		headerValue = bearerPrefix + headerValue
	}

	return setRequestHeader(req, serviceAuthTokenHeader, headerValue)
}

// SetDownloadServiceToken set the download service auth token header on the provided request. If the authentication
// token is already present it will be overwritten by the new value. If the header value is empty returns ErrValueEmpty
func SetDownloadServiceToken(req *http.Request, headerValue string) error {
	return setRequestHeader(req, downloadServiceTokenHeader, headerValue)
}

// SetUserIdentity set the user identity header on the provided request. If a user identity token is already present it
// will be overwritten by the new value. If the header value is empty returns ErrValueEmpty
func SetUserIdentity(req *http.Request, headerValue string) error {
	return setRequestHeader(req, userIdentityHeader, headerValue)
}

// SetRequestID set the unique request ID header on the provided request. If a request ID header is already present it
// will be overwritten by the new value. If the header value is empty returns ErrValueEmpty
func SetRequestID(req *http.Request, headerValue string) error {
	return setRequestHeader(req, requestIDHeader, headerValue)
}

// SetLocaleCode set the locale code header on the provided request. If this header is already present it
// will be overwritten by the new value. If the header value is empty returns ErrValueEmpty
func SetLocaleCode(req *http.Request, headerValue string) error {
	return setRequestHeader(req, localeCodeHeader, headerValue)
}

func setRequestHeader(req *http.Request, headerName string, headerValue string) error {
	if req == nil {
		return ErrRequestNil
	}

	if len(headerValue) == 0 {
		return ErrValueEmpty
	}

	req.Header.Set(headerName, headerValue)
	return nil
}
