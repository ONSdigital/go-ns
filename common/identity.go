package common

import (
	"context"
	"math/rand"
	"net/http"
	"time"
)

// ContextKey is an alias of type string
type ContextKey string

// A list of common constants used across go-ns packages
const (
	FlorenceHeaderKey        = "X-Florence-Token"
	DownloadServiceHeaderKey = "X-Download-Service-Token"

	FlorenceCookieKey = "access_token"

	AuthHeaderKey    = "Authorization"
	UserHeaderKey    = "User-Identity"
	RequestHeaderKey = "X-Request-Id"

	DeprecatedAuthHeader = "Internal-Token"
	LegacyUser           = "legacyUser"
	BearerPrefix         = "Bearer "

	UserIdentityKey   = ContextKey("User-Identity")
	CallerIdentityKey = ContextKey("Caller-Identity")
	RequestIdKey      = ContextKey("request-id")
)

// CheckRequester is an interface to allow mocking of auth.CheckRequest
type CheckRequester interface {
	CheckRequest(*http.Request) (context.Context, int, error)
}

// IdentityResponse represents the response from the identity service
type IdentityResponse struct {
	Identifier string `json:"identifier"`
}

// IsUserPresent determines if a user identity is present on the given context
func IsUserPresent(ctx context.Context) bool {
	userIdentity := ctx.Value(UserIdentityKey)
	return userIdentity != nil && userIdentity != ""

}

// AddUserHeader sets the given user ID on the given request
func AddUserHeader(r *http.Request, user string) {
	r.Header.Add(UserHeaderKey, user)
}

// AddServiceTokenHeader sets the given service token on the given request
func AddServiceTokenHeader(r *http.Request, serviceToken string) {
	if len(serviceToken) > 0 {
		r.Header.Add(AuthHeaderKey, BearerPrefix+serviceToken)
	}
}

// AddDownloadServiceTokenHeader sets the given download service token on the given request
func AddDownloadServiceTokenHeader(r *http.Request, serviceToken string) {
	if len(serviceToken) > 0 {
		r.Header.Add(DownloadServiceHeaderKey, serviceToken)
	}
}

// User gets the user identity from the context
func User(ctx context.Context) string {
	userIdentity, _ := ctx.Value(UserIdentityKey).(string)
	return userIdentity
}

// SetUser sets the user identity on the context
func SetUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, UserIdentityKey, user)
}

// AddAuthHeaders sets authentication headers for request
func AddAuthHeaders(ctx context.Context, r *http.Request, serviceToken string) {
	if IsUserPresent(ctx) {
		AddUserHeader(r, User(ctx))
	}
	AddServiceTokenHeader(r, serviceToken)
}

// AddDeprecatedHeader sets the deprecated header on the given request
func AddDeprecatedHeader(r *http.Request, token string) {
	if len(token) > 0 {
		r.Header.Add(DeprecatedAuthHeader, token)
	}
}

// IsCallerPresent determines if an identity is present on the given context.
func IsCallerPresent(ctx context.Context) bool {

	callerIdentity := ctx.Value(CallerIdentityKey)
	isPresent := callerIdentity != nil && callerIdentity != ""

	return isPresent
}

// Caller gets the caller identity from the context
func Caller(ctx context.Context) string {

	callerIdentity, _ := ctx.Value(CallerIdentityKey).(string)
	return callerIdentity
}

// SetCaller sets the caller identity on the context
func SetCaller(ctx context.Context, caller string) context.Context {

	return context.WithValue(ctx, CallerIdentityKey, caller)
}

// GetRequestId gets the correlation id on the context
func GetRequestId(ctx context.Context) string {
	correlationId, _ := ctx.Value(RequestIdKey).(string)
	return correlationId
}

// WithRequestId sets the correlation id on the context
func WithRequestId(ctx context.Context, correlationId string) context.Context {
	return context.WithValue(ctx, RequestIdKey, correlationId)
}

// AddRequestIdHeader add header for given correlation ID
func AddRequestIdHeader(r *http.Request, token string) {
	if len(token) > 0 {
		r.Header.Add(RequestHeaderKey, token)
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var requestIDRandom = rand.New(rand.NewSource(time.Now().UnixNano()))

// NewRequestID generates a random string of requested length
func NewRequestID(size int) string {
	b := make([]rune, size)
	for i := range b {
		b[i] = letters[requestIDRandom.Intn(len(letters))]
	}
	return string(b)
}
