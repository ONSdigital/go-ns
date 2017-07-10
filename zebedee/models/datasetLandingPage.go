package models

import "github.com/ONSdigital/dp-frontend-models/model"

// DatasetLandingPage is the page model of the Zebedee response for a dataset landing page type
type DatasetLandingPage struct {
	Type                      string          `json:"type"`
	URI                       string          `json:"uri"`
	Description               Description     `json:"description"`
	FilterID                  string          `json:"filterId"`
	Section                   Section         `json:"section"`
	Datasets                  []model.Related `json:"datasets"`
	RelatedLinks              []model.Related `json:"links"`
	RelatedDatasets           []model.Related `json:"relatedDatasets"`
	RelatedDocuments          []model.Related `json:"relatedDocuments"`
	RelatedMethodology        []model.Related `json:"relatedMethodology"`
	RelatedMethodologyArticle []model.Related `json:"relatedMethodologyArticle"`
	Alerts                    []Alert         `json:"alerts"`
	Timeseries                bool            `json:"timeseries"`
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
	Email     string `json:"email"`
	Name      string `json:"name"`
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
