package request

import (
	"github.com/ONSdigital/go-ns/log"
	"io"
	"io/ioutil"
	"net/http"
	"github.com/pkg/errors"
)

// DrainBody drains the body of the given of the given HTTP request.
func DrainBody(r *http.Request) {
	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		log.ErrorCtx(r.Context(), errors.Wrap(err, "error draining request body"), nil)
	}

	err = r.Body.Close()
	if err != nil {
		log.ErrorCtx(r.Context(), errors.Wrap(err, "error closing request body"), nil)
	}
}
