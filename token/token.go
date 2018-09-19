package token

import (
	"errors"
	"net/http"
)

const (
	identityToken = "token"
)

// GetToken extracts a token from a given request
func GetToken(r *http.Request) (string, error) {
	token := r.Header.Get(identityToken)
	if token == "" {
		return "", errors.New("No token found in request header")
	}
	return token, nil
}
