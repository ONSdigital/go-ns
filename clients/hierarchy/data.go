package hierarchy

// Model represents the model returned by the heirarchy api
type Model struct {
	ID       string  `json:"id"`
	Label    string  `json:"label"`
	Children []Child `json:"children"`
	Parent   Parent  `json:"parent"`
}

// Child represents a child item in the hierarchy model
type Child struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	URL              string `json:"url"`
	NumberofChildren int    `json:"number_of_children"`
	LabelCode        string `json:"label_code"`
}

// Parent represents a parent item in the hierarchy model
type Parent struct {
	URL       string `json:"url"`
	Label     string `json:"label"`
	LabelCode string `json:"label_code"`
	ID        string `json:"id"`
}
