package dataset

// Model represents a response model from the dataset api
type Model struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	ReleaseDate string  `json:"release_date"`
	NextRelease string  `json:"next_release"`
	Edition     string  `json:"edition"`
	Version     string  `json:"version"`
	Contact     Contact `json:"contact"`
}

// Contact represents a response model within a dataset
type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

// Metadata represents metadata from dataset
type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
