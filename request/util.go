package request

import (
	"github.com/ONSdigital/go-ns/log"
	"io"
	"io/ioutil"
	"net/http"
)

// DrainBody drains the body of the given of the given HTTP request.
// any handler function that reads the request body should ensure that this
// function is called.
// ``
func DrainBody(r *http.Request) {
	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		log.InfoCtx(r.Context(), "error draining request body", nil)
	}

	err = r.Body.Close()
	if err != nil {
		log.InfoCtx(r.Context(), "error closing request body", nil)
	}
}
