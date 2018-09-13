package models

// EditionsListResults contains an array of code lists which can be paginated
type EditionsListResults struct {
	Items      []EditionsList `json:"items"`
	Count      int            `json:"count"`
	Offset     int            `json:"offset"`
	Limit      int            `json:"limit"`
	TotalCount int            `json:"total_count"`
}

// EditionsList containing links to all possible codes
type EditionsList struct {
	Edition string           `json:"edition"`
	Label   string           `json:"label"`
	Links   EditionsListLink `json:"links"`
}

// EditionsListLink contains links for a code list resource
type EditionsListLink struct {
	Self     *Link `json:"self"`
	Editions *Link `json:"editions"`
	Codes    *Link `json:"codes"`
}
