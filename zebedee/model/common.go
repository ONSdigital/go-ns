package model

// PageDescription is a common section for every page containing common fields.
type PageDescription struct {
	Title       string   `json:"title"`
	Summary     string   `json:"description"`
	Keywords    []string `json:"keywords"`
	ReleaseDate string   `json:"releaseDate"`
	PreUnit     string   `json:"preUnit"`
	Unit        string   `json:"unit"`
	Number      string   `json:"number"`
}
