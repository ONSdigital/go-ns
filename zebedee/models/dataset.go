package models

// Dataset represents a datase response from zebedee
type Dataset struct {
	Type               string              `json:"type"`
	URI                string              `json:"uri"`
	Description        Description         `json:"description"`
	Downloads          []Download          `json:"downloads"`
	SupplementaryFiles []SupplementaryFile `json:"supplementaryFiles"`
	Versions           []Version           `json:"versions"`
}

// Download represents download within a dataset
type Download struct {
	File string `json:"file"`
}

// SupplementaryFile represents a SupplementaryFile within a dataset
type SupplementaryFile struct {
	Title string `json:"title"`
	File  string `json:"file"`
}

// Version represents a version of a dataset
type Version struct {
	URI         string `json:"uri"`
	ReleaseDate string `json:"updateDate"`
	Notice      string `json:"correctionNotice"`
	Label       string `json:"label"`
}
