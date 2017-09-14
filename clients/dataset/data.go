package dataset

// Model represents a response dataset model from the dataset api
type Model struct {
	CollectionID string    `json:"collection_id"`
	Contact      Contact   `json:"contact"`
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
	Dataset       Link `json:"dataset"`
	Dimensions    Link `json:"dimensions"`
	Edition       Link `json:"edition"`
	Editions      Link `json:"editions"`
	LatestVersion Link `json:"latest_version"`
	Versions      Link `json:"versions"`
	Self          Link `json:"self"`
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
