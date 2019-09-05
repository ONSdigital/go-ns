package clients

// AuthToken defines common behaviour for authentication tokens
type AuthToken interface {
	GetValue() string
}

// UserAuthToken is a type representing user authentication tokens
type UserAuthToken struct {
	value string
}

// ServiceAuthToken is a type representing user authentication tokens
type ServiceAuthToken struct {
	value string
}

// NewUserAuthToken returns a NewUserAuthToken from the token string provided. If the string is empty then nil is returned
func NewUserAuthToken(token string) *UserAuthToken {
	var authToken *UserAuthToken

	if len(token) > 0 {
		authToken = &UserAuthToken{
			value: token,
		}
	}
	return authToken
}

// GetValue returns the token string value if the token is not nil, otherwise returns an empty string
func (token *UserAuthToken) GetValue() string {
	if token == nil {
		return ""
	}
	return token.value
}

// NewServiceAuthToken returns a ServiceAuthToken from the token string provided. If the string is empty then nil is returned
func NewServiceAuthToken(token string) *ServiceAuthToken {
	var authToken *ServiceAuthToken

	if len(token) > 0 {
		authToken = &ServiceAuthToken{
			value: token,
		}
	}
	return authToken
}

// GetValue returns the token string value if the token is not nil, otherwise returns an empty string
func (token *ServiceAuthToken) GetValue() string {
	if token == nil {
		return ""
	}
	return token.value
}
