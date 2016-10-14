package healthcheck

import "net/http"

func Handler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
}
