package data

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
	Size string
}

// FileSize represents a file size from zebedee
type FileSize struct {
	Size int `json:"fileSize"`
}

// PageTitle represents a page title from zebedee
type PageTitle struct {
	Title   string `json:"title"`
	Edition string `json:"edition"`
}

// SupplementaryFile represents a SupplementaryFile within a dataset
type SupplementaryFile struct {
	Title string `json:"title"`
	File  string `json:"file"`
	Size  string
}

// Version represents a version of a dataset
type Version struct {
	URI         string `json:"uri"`
	ReleaseDate string `json:"updateDate"`
	Notice      string `json:"correctionNotice"`
	Label       string `json:"label"`
}

// Breadcrumb represents a breadcrumb from zebedee
type Breadcrumb struct {
	URI         string          `json:"uri"`
	Description NodeDescription `json:"description"`
	Type        string          `json:"type"`
}

// NodeDescription represents a node description
type NodeDescription struct {
	Title string `json:"title"`
}

// DatasetLandingPage is the page model of the Zebedee response for a dataset landing page type
type DatasetLandingPage struct {
	Type                      string      `json:"type"`
	URI                       string      `json:"uri"`
	Description               Description `json:"description"`
	Section                   Section     `json:"section"`
	Datasets                  []Related   `json:"datasets"`
	RelatedLinks              []Related   `json:"links"`
	RelatedFilterableDatasets []Related   `json:"relatedFilterableDatasets"`
	RelatedDatasets           []Related   `json:"relatedDatasets"`
	RelatedDocuments          []Related   `json:"relatedDocuments"`
	RelatedMethodology        []Related   `json:"relatedMethodology"`
	RelatedMethodologyArticle []Related   `json:"relatedMethodologyArticle"`
	Alerts                    []Alert     `json:"alerts"`
	Timeseries                bool        `json:"timeseries"`
}

// Description represents a description block within the dataset landing page
type Description struct {
	Title             string   `json:"title"`
	Edition           string   `json:"edition"`
	Summary           string   `json:"summary"`
	Keywords          []string `json:"keywords"`
	MetaDescription   string   `json:"metaDescription"`
	NationalStatistic bool     `json:"nationalStatistic"`
	Contact           Contact  `json:"contact"`
	ReleaseDate       string   `json:"releaseDate"`
	NextRelease       string   `json:"nextRelease"`
	DatasetID         string   `json:"datasetId"`
	Unit              string   `json:"unit"`
	PreUnit           string   `json:"preUnit"`
	Source            string   `json:"source"`
}

// Contact represents a contact within dataset landing page
type Contact struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Telephone string `json:"telephone"`
}

// Section represents a section within dataset landing page
type Section struct {
	Markdown string `json:"markdown"`
}

// Alert represents an alert within dataset landing page
type Alert struct {
	Date     string `json:"date"`
	Markdown string `json:"markdown"`
	Type     string `json:"type"`
}

//Related stores the Title and URI for any related data (eg related publications on a dataset page)
type Related struct {
	Title string `json:"title"`
	URI   string `json:"uri"`
}
