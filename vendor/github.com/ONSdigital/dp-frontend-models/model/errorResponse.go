package model

//ErrorResponse store error from JSON unmarshalling so it can be returned as a response from API
type ErrorResponse struct {
	Error string `json:"error"`
}
