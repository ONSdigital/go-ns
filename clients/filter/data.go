package filter

// Dimension represents a dimension response from the filter api
type Dimension struct {
	Name string `json:"name"`
	URI  string `json:"dimension_url"`
}

// DimensionOption represents a dimension option from the filter api
type DimensionOption struct {
	DimensionOptionsURL string `json:"dimension_option_url"`
	Option              string `json:"option"`
}

// Model represents a model returned from the filter api
type Model struct {
	FilterID        string              `json:"filter_job_id"`
	DatasetFilterID string              `json:"dataset_filter_id"`
	Dataset         string              `json:"dataset"`
	Edition         string              `json:"edition"`
	Version         string              `json:"version"`
	State           string              `json:"state"`
	Dimensions      []ModelDimension    `json:"dimensions"`
	Downloads       map[string]Download `json:"downloads"`
	Events          map[string][]Event  `json:"events"`
}

// ModelDimension represents a dimension to be filtered upon
type ModelDimension struct {
	Name      string    `json:"name"`
	Values    []string  `json:"values"`
	IDs       []string  `json:"ids"`
	Hierarchy Hierarchy `json:"hierarchy"`
}

// Hierarchy represents a hierarchy in a filter model
type Hierarchy struct {
	ID string `json:"id"`
}

// Download represents a download within a filter from api response
type Download struct {
	URL  string `json:"url"`
	Size string `json:"size"`
}

// Event represents an event from a filter api response
type Event struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Message string `json:"message"`
}
