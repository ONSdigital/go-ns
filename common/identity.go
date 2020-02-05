package common

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ContextKey is an alias of type string
type ContextKey string

// A list of common constants used across go-ns packages
const (
	FlorenceCookieKey = "access_token"

	UserIdentityKey     = ContextKey("User-Identity")
	CallerIdentityKey   = ContextKey("Caller-Identity")
	RequestIdKey        = ContextKey("request-id")
	FlorenceIdentityKey = ContextKey("florence-id")
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

// IsFlorenceIdentityPresent determines if a florence identity is present on the given context
func IsFlorenceIdentityPresent(ctx context.Context) bool {
	florenceID := ctx.Value(FlorenceIdentityKey)
	return florenceID != nil && florenceID != ""
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

// SetFlorenceIdentity sets the florence identity for authentication
func SetFlorenceIdentity(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, FlorenceIdentityKey, user)
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

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var requestIDRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
var randMutex sync.Mutex

// NewRequestID generates a random string of requested length
func NewRequestID(size int) string {
	b := make([]rune, size)
	randMutex.Lock()
	for i := range b {
		b[i] = letters[requestIDRandom.Intn(len(letters))]
	}
	randMutex.Unlock()
	return string(b)
}
