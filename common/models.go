package common

// Params represents a generic map of key value pairs
type Params map[string]string

// Copy preserves the original params value (key value pair)
// but stores the data in a different reference address
func (originalParams Params) Copy() Params {
	if originalParams == nil {
		return nil
	}

	params := Params{}
	for key, value := range originalParams {
		params[key] = value
	}

	return params
}
