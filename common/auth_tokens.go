package common

import (
	"net/http"
)

// AuthToken defines common behaviour for authentication tokens
type AuthToken interface {
	GetValue() string
	SetAsRequestHeader(req *http.Request)
}

// UserAuthToken is a type representing user authentication tokens
type UserAuthToken struct {
	value string
}

// ServiceAuthToken is a type representing user authentication tokens
type ServiceAuthToken struct {
	value string
}

// NewUserAuthToken returns a NewUserAuthToken from the token string provided. If the string is empty then nil is returned
func NewUserAuthToken(token string) *UserAuthToken {
	var authToken *UserAuthToken

	if len(token) > 0 {
		authToken = &UserAuthToken{
			value: token,
		}
	}
	return authToken
}

// GetValue returns the token string value if the token is not nil, otherwise returns an empty string
func (token *UserAuthToken) GetValue() string {
	if token == nil {
		return ""
	}
	return token.value
}

// SetAsRequestHeader set the user auth token as a request header. Do nothing If the token is nil or empty or if the request is nil.
func (token *UserAuthToken) SetAsRequestHeader(req *http.Request) {
	if req != nil && token != nil && len(token.GetValue()) > 0 {
		req.Header.Set(FlorenceHeaderKey, token.GetValue())
	}
}

// NewServiceAuthToken returns a ServiceAuthToken from the token string provided. If the string is empty then nil is returned
func NewServiceAuthToken(token string) *ServiceAuthToken {
	var authToken *ServiceAuthToken

	if len(token) > 0 {
		authToken = &ServiceAuthToken{
			value: token,
		}
	}
	return authToken
}

// GetValue returns the token string value if the token is not nil, otherwise returns an empty string
func (token *ServiceAuthToken) GetValue() string {
	if token == nil {
		return ""
	}
	return token.value
}

// SetAsRequestHeader set the service auth token as a request header. Do nothing If the token is nil or empty or if the request is nil.
func (token *ServiceAuthToken) SetAsRequestHeader(req *http.Request) {
	if req != nil && token != nil && len(token.GetValue()) > 0 {
		req.Header.Set(AuthHeaderKey, BearerPrefix+token.GetValue())
	}
}
