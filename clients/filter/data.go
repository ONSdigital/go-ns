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

// CreateBlueprint represents the fields required to create a filter blueprint
type CreateBlueprint struct {
	Dataset    Dataset          `json:"dataset"`
	Dimensions []ModelDimension `json:"dimensions"`
	FilterID   string           `json:"filter_id"`
}

// Dataset represents the dataset fields required to create a filter blueprint
type Dataset struct {
	DatasetID string `json:"id"`
	Edition   string `json:"edition"`
	Version   int    `json:"version"`
}

// Model represents a model returned from the filter api
type Model struct {
	FilterID    string              `json:"filter_id"`
	InstanceID  string              `json:"instance_id"`
	Links       Links               `json:"links"`
	DatasetID   string              `json:"dataset_id"`
	Edition     string              `json:"edition"`
	Version     string              `json:"version"`
	State       string              `json:"state"`
	Dimensions  []ModelDimension    `json:"dimensions,omitempty"`
	Downloads   map[string]Download `json:"downloads,omitempty"`
	Events      []Event             `json:"events,omitempty"`
	IsPublished bool                `json:"published"`
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
	Name    string   `json:"name"`
	Options []string `json:"options"`
	Values  []string `json:"values"`
}

// Download represents a download within a filter from api response
type Download struct {
	URL     string `json:"href"`
	Size    string `json:"size"`
	Public  string `json:"public,omitempty"`
	Private string `json:"private,omitempty"`
	Skipped bool   `json:"skipped,omitempty"`
}

// Event represents an event from a filter api response
type Event struct {
	Time string `json:"time"`
	Type string `json:"type"`
}

// Preview represents a preview document returned from the filter api
type Preview struct {
	Headers         []string   `json:"headers"`
	NumberOfRows    int        `json:"number_of_rows"`
	NumberOfColumns int        `json:"number_of_columns"`
	Rows            [][]string `json:"rows"`
}
