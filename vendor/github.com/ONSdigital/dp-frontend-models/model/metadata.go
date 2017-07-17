package model

//Metadata ...
type Metadata struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Footer      Footer   `json:"footer"`
}

// Footer ...
type Footer struct {
	Enabled     bool   `json:"enabled"`
	Contact     string `json:"contact"`
	ReleaseDate string `json:"release_date"`
	NextRelease string `json:"next_release"`
	DatasetID   string `json:"dataset_id"`
}
