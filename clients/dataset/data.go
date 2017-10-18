package dataset

// Model represents a response dataset model from the dataset api
type Model struct {
	CollectionID string    `json:"collection_id"`
	Contacts     []Contact `json:"contacts"`
	Description  string    `json:"description"`
	Links        Links     `json:"links"`
	NextRelease  string    `json:"next_release"`
	Periodicity  string    `json:"yearly"`
	Publisher    Publisher `json:"publisher"`
	State        string    `json:"state"`
	Theme        string    `json:"theme"`
	Title        string    `json:"title"`
}

// Version represents a version within a dataset
type Version struct {
	CollectionID string `json:"collection_id"`
	Edition      string `json:"edition"`
	ID           string `json:"id"`
	InstanceID   string `json:"instance_id"`
	License      string `json:"license"`
	Links        Links  `json:"links"`
	ReleaseDate  string `json:"release_date"`
	State        string `json:"date"`
	Version      int    `json:"version"`
}

// Edition represents an edition within a dataset
type Edition struct {
	Edition string `json:"edition"`
	ID      string `json:"id"`
	Links   Links  `json:"links"`
	State   string `json:"state"`
}

// Publisher represents the publisher within the dataset
type Publisher struct {
	URL  string `json:"href"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Links represent the Links within a dataset model
type Links struct {
	Dataset       Link `json:"dataset,omitempty"`
	Dimensions    Link `json:"dimensions,omitempty"`
	Edition       Link `json:"edition,omitempty"`
	Editions      Link `json:"editions,omitempty"`
	LatestVersion Link `json:"latest_version,omitempty"`
	Versions      Link `json:"versions,omitempty"`
	Self          Link `json:"self,omitempty"`
	CodeList      Link `json:"code_list,omitempty"`
	Options       Link `json:"options,omitempty"`
	Version       Link `json:"version,omitempty"`
	Code          Link `json:"code,omitempty"`
}

// Link represents a single link within a dataset model
type Link struct {
	URL string `json:"href"`
	ID  string `json:"id,omitempty"`
}

// Contact represents a response model within a dataset
type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

// Dimensions represent a list of dimensions from the dataset api
type Dimensions struct {
	Items []Dimension `json:"items"`
}

// Dimension represents a response model for a dimension endpoint
type Dimension struct {
	ID    string `json:"dimension_id"`
	Links Links  `json:"links"`
}

// Options represents a list of options from the dataset api
type Options struct {
	Items []Option `json:"items"`
}

// Option represents a response model for an option
type Option struct {
	DimensionID string `json:"dimension_id"`
	Label       string `json:"label"`
	Links       Links  `json:"links"`
	Option      string `json:"option"`
}
