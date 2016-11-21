package common

type ONSError struct {
	Parameters map[string]interface{}
	RootError  error
}

func NewONSError(e error, description string) *ONSError {
	err := &ONSError{RootError: e}
	err.AddParameter("ErrorDescription", description)
	return err
}

func (e ONSError) Error() string {
	return e.RootError.Error()
}

func (e *ONSError) AddParameter(name string, value interface{}) *ONSError {
	if e.Parameters == nil {
		e.Parameters = make(map[string]interface{}, 0)
	}
	e.Parameters[name] = value
	return e
}
