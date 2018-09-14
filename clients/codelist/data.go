package codelist

// DimensionValues represent the dimension values returned by the codelist api
type DimensionValues struct {
	Items           []Item `json:"items"`
	NumberOfResults int    `json:"number_of_results"`
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

// Link contains the id and a link to a resource
type Link struct {
	ID   string `json:"id,omitempty"     bson:"id"`
	Href string `json:"href"             bson:"href"`
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

// CodesResults contains the list of codes for a specific code list and edition
type CodesResults struct {
	Items      []Item `json:"items"`
	Count      int    `json:"count"`
	Offset     int    `json:"offset"`
	Limit      int    `json:"limit"`
	TotalCount int    `json:"total_count"`
}

// Item represents an individual code item returned by the codelist api
type Item struct {
	ID    string    `json:"id"`
	Label string    `json:"label"`
	Links CodeLinks `json:"links"`
}

// CodeLinks represents the links an individual code item has
type CodeLinks struct {
	CodeLists Link `json:"code_lists"`
	Datasets  Link `json:"datasets"`
	Self      Link `json:"self"`
}
