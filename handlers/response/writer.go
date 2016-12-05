package response

import (
	"net/http"
	"encoding/json"
)

const contentTypeHeader = "Content-Type"
const contentTypeJSON = "application/json"

// JSONEncoder interface defining a JSON encoder.
type JSONEncoder interface {
	writeResponseJSON(w http.ResponseWriter, value interface{}, status int) error
}

type onsJSONEncoder struct{}

var jsonResponseEncoder JSONEncoder = &onsJSONEncoder{}

// WriteJSON set the content type header to JSON, writes the response object as json and sets the http status code.
func WriteJSON(w http.ResponseWriter, value interface{}, status int) error {
	return jsonResponseEncoder.writeResponseJSON(w, value, status);
}

func (j *onsJSONEncoder) writeResponseJSON(w http.ResponseWriter, value interface{}, status int) error {
	w.Header().Set(contentTypeHeader, contentTypeJSON)

	b, err := json.Marshal(value)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	w.WriteHeader(status)
	w.Write(b)
	return nil
}