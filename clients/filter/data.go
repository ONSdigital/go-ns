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
	FilterID   string              `json:"filter_id"`
	InstanceID string              `json:"instance_id"`
	Links      Links               `json:"links"`
	Dataset    string              `json:"dataset"`
	Edition    string              `json:"edition"`
	Version    string              `json:"version"`
	State      string              `json:"state"`
	Dimensions []ModelDimension    `json:"dimensions,omitempty"`
	Downloads  map[string]Download `json:"downloads,omitempty"`
	Events     map[string][]Event  `json:"events,omitempty"`
}

// Links represents a links object on the filter api response
type Links struct {
	Version         Link `json:"version,omitempty"`
	FilterOutputs   Link `json:"filter_output,omitempty"`
	FilterBlueprint Link `json:"filter_blueprint,omitempty"`
}

// Link represents a single link within a links object
type Link struct {
	ID   string `json:"id"`
	HRef string `json:"href"`
}

// ModelDimension represents a dimension to be filtered upon
type ModelDimension struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
	IDs    []string `json:"ids"`
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
