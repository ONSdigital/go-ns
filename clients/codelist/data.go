package codelist

// DimensionValues represent the dimension values returned by the codelist api
type DimensionValues struct {
	Items           []Item `json:"items"`
	NumberOfResults int    `json:"number_of_results"`
}

// Item represents an individual item returned by the codelist api
type Item struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Name  string `json:"name"`
}
