package dataset

import (
	"bytes"
	"fmt"
	"unicode"
)

// Model represents a response dataset model from the dataset api
type Model struct {
	ID                string           `json:"id"`
	CollectionID      string           `json:"collection_id"`
	Contacts          []Contact        `json:"contacts"`
	Description       string           `json:"description"`
	Keywords          []string         `json:"keywords"`
	License           string           `json:"license"`
	Links             Links            `json:"links"`
	Methodologies     []Methodology    `json:"methodologies"`
	NationalStatistic bool             `json:"national_statistic"`
	NextRelease       string           `json:"next_release"`
	Publications      []Publication    `json:"publications"`
	Publisher         Publisher        `json:"publisher"`
	QMI               Publication      `json:"qmi"`
	RelatedDatasets   []RelatedDataset `json:"related_datasets"`
	ReleaseFrequency  string           `json:"release_frequency"`
	State             string           `json:"state"`
	Theme             string           `json:"theme"`
	Title             string           `json:"title"`
	UnitOfMeasure     string           `json:"unit_of_measure"`
	URI               string           `json:"uri"`
}

// Version represents a version within a dataset
type Version struct {
	Alerts        []Alert             `json:"alerts"`
	CollectionID  string              `json:"collection_id"`
	Downloads     map[string]Download `json:"downloads"`
	Edition       string              `json:"edition"`
	ID            string              `json:"id"`
	InstanceID    string              `json:"instance_id"`
	LatestChanges []Change            `json:"latest_changes"`
	License       string              `json:"license"`
	Links         Links               `json:"links"`
	ReleaseDate   string              `json:"release_date"`
	State         string              `json:"date"`
	Temporal      []Temporal          `json:"temporal"`
	Version       int                 `json:"version"`
}

// Metadata is a combination of version and dataset model fields
type Metadata struct {
	Version
	Model
}

// Download represents a version download from the dataset api
type Download struct {
	URL  string `json:"url"`
	Size string `json:"size"`
}

// Edition represents an edition within a dataset
type Edition struct {
	Edition string `json:"edition"`
	ID      string `json:"id"`
	Links   Links  `json:"links"`
	State   string `json:"state"`
}

// Publisher represents the publisher within the dataset
type Publisher struct {
	URL  string `json:"href"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Links represent the Links within a dataset model
type Links struct {
	AccessRights  Link `json:"access_rights,omitempty"`
	Dataset       Link `json:"dataset,omitempty"`
	Dimensions    Link `json:"dimensions,omitempty"`
	Edition       Link `json:"edition,omitempty"`
	Editions      Link `json:"editions,omitempty"`
	LatestVersion Link `json:"latest_version,omitempty"`
	Versions      Link `json:"versions,omitempty"`
	Self          Link `json:"self,omitempty"`
	CodeList      Link `json:"code_list,omitempty"`
	Options       Link `json:"options,omitempty"`
	Version       Link `json:"version,omitempty"`
	Code          Link `json:"code,omitempty"`
}

// Link represents a single link within a dataset model
type Link struct {
	URL string `json:"href"`
	ID  string `json:"id,omitempty"`
}

// Contact represents a response model within a dataset
type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

// Dimensions represent a list of dimensions from the dataset api
type Dimensions struct {
	Items Items `json:"items"`
}

// Items represents a list of dimensions
type Items []Dimension

func (d Items) Len() int      { return len(d) }
func (d Items) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d Items) Less(i, j int) bool {
	iRunes := []rune(d[i].ID)
	jRunes := []rune(d[j].ID)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	return false
}

// Dimension represents a response model for a dimension endpoint
type Dimension struct {
	ID          string `json:"dimension"`
	Links       Links  `json:"links"`
	Description string `json:"description"`
}

// Options represents a list of options from the dataset api
type Options struct {
	Items []Option `json:"items"`
}

// Option represents a response model for an option
type Option struct {
	DimensionID string `json:"dimension"`
	Label       string `json:"label"`
	Links       Links  `json:"links"`
	Option      string `json:"option"`
}

// Methodology represents a methodology document returned by the dataset api
type Methodology struct {
	Description string `json:"description"`
	URL         string `json:"href"`
	Title       string `json:"title"`
}

// Publication represents a publication document returned by the dataset api
type Publication struct {
	Description string `json:"description"`
	URL         string `json:"href"`
	Title       string `json:"title"`
}

// RelatedDataset represents a related dataset document returned by the dataset api
type RelatedDataset struct {
	URL   string `json:"href"`
	Title string `json:"title"`
}

// Alert represents an alert returned by the dataset api
type Alert struct {
	Date        string `json:"date"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// Change represents a change returned for a version by the dataset api
type Change struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

// Temporal represents a temporal returned by the dataset api
type Temporal struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Frequency string `json:"frequency"`
}

func (m Metadata) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("Title: %s\n", m.Title))
	b.WriteString(fmt.Sprintf("Description: %s\n", m.Description))
	b.WriteString(fmt.Sprintf("Publisher: %s\n", m.Publisher))
	b.WriteString(fmt.Sprintf("Issued: %s\n", m.ReleaseDate))
	b.WriteString(fmt.Sprintf("Next Release: %s\n", m.NextRelease))
	b.WriteString(fmt.Sprintf("Identifier: %s\n", m.Title))
	b.WriteString(fmt.Sprintf("Keywords: %s\n", m.Keywords))
	b.WriteString(fmt.Sprintf("Language: %s\n", "English"))
	if len(m.Contacts) > 0 {
		b.WriteString(fmt.Sprintf("Contact: %s, %s, %s\n", m.Contacts[0].Name, m.Contacts[0].Email, m.Contacts[0].Telephone))
	}
	if len(m.Temporal) > 0 {
		b.WriteString(fmt.Sprintf("Temporal: %s\n", m.Temporal[0].Frequency))
	}
	b.WriteString(fmt.Sprintf("Latest Changes: %s\n", m.LatestChanges))
	b.WriteString(fmt.Sprintf("Periodicity: %s\n", m.ReleaseFrequency))
	b.WriteString(fmt.Sprintf("Distribution: %s\n", m.Downloads))
	b.WriteString(fmt.Sprintf("Unit of measure: %s\n", m.UnitOfMeasure))
	b.WriteString(fmt.Sprintf("License: %s\n", m.Model.License))
	b.WriteString(fmt.Sprintf("Methodologies: %s\n", m.Methodologies))
	b.WriteString(fmt.Sprintf("National Statistic: %t\n", m.NationalStatistic))
	b.WriteString(fmt.Sprintf("Publications: %s\n", m.Publications))
	b.WriteString(fmt.Sprintf("Related Links: %s\n", m.RelatedDatasets))

	return b.String()
}

func (m Options) String() string {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("\n\tTitle: %s\n", m.Items[0].DimensionID))
	var labels, options []string

	for _, dim := range m.Items {
		labels = append(labels, dim.Label)
		options = append(options, dim.Option)
	}

	b.WriteString(fmt.Sprintf("\tLabels: %s\n", labels))
	b.WriteString(fmt.Sprintf("\tOptions: %v\n", options))

	return b.String()
}
