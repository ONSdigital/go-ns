package request

import (
	"github.com/ONSdigital/go-ns/log"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
)

// DrainBody drains the body of the given of the given HTTP request.
func DrainBody(r *http.Request) {

	if r.Body == nil {
		return
	}

	_, err := io.Copy(ioutil.Discard, r.Body)
	if err != nil {
		log.ErrorCtx(r.Context(), errors.Wrap(err, "error draining request body"), nil)
	}

	err = r.Body.Close()
	if err != nil {
		log.ErrorCtx(r.Context(), errors.Wrap(err, "error closing request body"), nil)
	}
}
