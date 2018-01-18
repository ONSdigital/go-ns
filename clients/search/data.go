package search

// Model represents a model returned by the search api
type Model struct {
	Count      int    `json:"count"`
	Items      []Item `json:"items"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	TotalCount int    `json:"total_count"`
}

// Item represents a single hierarchy item returned by the search api
type Item struct {
	Code               string  `json:"code"`
	DimensionOptionURL string  `json:"dimension_option_url"`
	HasData            bool    `json:"has_data"`
	Label              string  `json:"label"`
	Matches            Matches `json:"matches"`
	NumberOfChildren   int     `json:"number_of_children"`
}

// Matches represent matches from the input query against the returned item
type Matches struct {
	Code  []Match `json:"code"`
	Label []Match `json:"label"`
}

// Match defines the start and end character numbers that the item matched with
type Match struct {
	Start int `json:"start"`
	End   int `json:"end"`
}
