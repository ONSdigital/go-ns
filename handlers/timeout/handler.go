package timeout

import (
	"net/http"
	"time"
)

// Handler implements a HTTP timeout
func Handler(timeout time.Duration) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.TimeoutHandler(h, timeout, "timed out")
	}
}
