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

// Link contains the id and a link to a resource
type Link struct {
	ID   string `json:"id,omitempty"     bson:"id"`
	Href string `json:"href"             bson:"href"`
}

// CodeListResults contains an array of code lists which can be paginated
type CodeListResults struct {
	Items      []CodeList `json:"items"`
	Count      int        `json:"count"`
	Offset     int        `json:"offset"`
	Limit      int        `json:"limit"`
	TotalCount int        `json:"total_count"`
}

// CodeList containing links to all possible codes
type CodeList struct {
	Links CodeListLinks `json:"links"`
}

// CodeListLinks contains links for a code list resource
type CodeListLinks struct {
	Self     *Link `json:"self"`
	Editions *Link `json:"editions"`
}

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
