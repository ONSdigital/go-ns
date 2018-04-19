package common

import (
	"context"
	"net/http"
)

type ContextKey string

const (
	FlorenceHeaderKey        = "X-Florence-Token"
	DownloadServiceHeaderKey = "X-Download-Service-Token"

	AuthHeaderKey = "Authorization"
	UserHeaderKey = "User-Identity"

	DeprecatedAuthHeader = "Internal-Token"
	LegacyUser           = "legacyUser"
	BearerPrefix         = "Bearer "

	UserIdentityKey   = ContextKey("User-Identity")
	CallerIdentityKey = ContextKey("Caller-Identity")
)

// interface to allow mocking of auth.CheckRequest
type CheckRequester interface {
	CheckRequest(*http.Request) (context.Context, int, error)
}

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
