package model

// HomePage is the root structure of the homepage
type HomePage struct {
	Type        string          `json:"type"`
	URI         string          `json:"uri"`
	Sections    []*HomeSection  `json:"sections"`
	Description PageDescription `json:"description"`
	Taxonomy    []ContentNode   `json:"taxonomy"`
}

// HomeSection represents a single section of the homepage.
type HomeSection struct {
	Index      int   `json:"index"`
	Theme      *Link `json:"theme"`
	Statistics *Link `json:"statistics"`
}

// Link represents a single link contained in a page.
type Link struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
	Index int    `json:"index"`
}
