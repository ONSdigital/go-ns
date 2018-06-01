package common

// ONSError encapsulates an error with additional parameters
type ONSError struct {
	Parameters map[string]interface{}
	RootError  error
}

// NewONSError creates a new ONS error
func NewONSError(e error, description string) *ONSError {
	err := &ONSError{RootError: e}
	err.AddParameter("ErrorDescription", description)
	return err
}

// Error returns the ONSError RootError message
func (e ONSError) Error() string {
	return e.RootError.Error()
}

// AddParameter method creates or overwrites parameters attatched to an ONSError
func (e *ONSError) AddParameter(name string, value interface{}) *ONSError {
	if e.Parameters == nil {
		e.Parameters = make(map[string]interface{}, 0)
	}
	e.Parameters[name] = value
	return e
}
