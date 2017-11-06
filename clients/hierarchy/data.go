package hierarchy

// Model represents the model returned by the heirarchy api
type Model struct {
	Label       string       `json:"label"`
	Links       Links        `json:"links"`
	Children    []Child      `json:"children,omitempty"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs,omitempty"`
}

// Links represents links within the hierarchy api
type Links struct {
	Self Link `json:"self"`
	Code Link `json:"code"`
}

// Link represents a link within the hierarchy api
type Link struct {
	ID  string `json:"id"`
	URL string `json:"href"`
}

// Child represents a child item in the hierarchy model
type Child struct {
	Label            string `json:"label"`
	NumberofChildren int    `json:"no_of_children,omitempty"`
	HasData          bool   `json:"has_data"`
	Links            Links  `json:"links"`
}

// Breadcrumb represents a breadcrumb item in the hierarchy model
type Breadcrumb struct {
	Label            string `json:"label"`
	NumberofChildren int    `json:"no_of_children,omitempty"`
	HasData          bool   `json:"has_data"`
	Links            Links  `json:"links"`
}
