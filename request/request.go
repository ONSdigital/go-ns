package request

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/log.go/v2/log"
)

// DrainBody drains the body of the given of the given HTTP request.
func DrainBody(r *http.Request) {

	if r.Body == nil {
		return
	}

	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		log.Error(r.Context(), "error draining request body", err)
	}

	err = r.Body.Close()
	if err != nil {
		log.Error(r.Context(), "error closing request body", err)
	}
}
