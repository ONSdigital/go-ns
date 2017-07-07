package zebedee

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ONSdigital/dp-frontend-models/model"
	"github.com/ONSdigital/dp-frontend-models/model/datasetLandingPageStatic"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/zebedee/models"
	"github.com/c2h5oh/datasize"
)

// Client is an interface for communication to a zebedee server:
// https://github.com/ONSdigital/zebedee
//
// Get will return a response from the server as a byte array, usually in a json
// format.
//
// GetLanding will return a static dataset landing page struct populated with the
// data from a given path in zebedee.
type Client interface {
	Get(path string) ([]byte, error)
	GetLanding(path string) (StaticDatasetLandingPage, error)
	SetAccessToken(token string)
}

type ZebedeeClient struct {
	zebedeeURL  string
	client      *http.Client
	accessToken string
}

// StaticDatasetLandingPage is a StaticDatasetLandingPage representation
type StaticDatasetLandingPage datasetLandingPageStatic.Page

// ErrInvalidZebedeeResponse is returned when zebedee does not respond
// with a valid status
type ErrInvalidZebedeeResponse struct {
	err  error
	data zebedeeRequestErrorData
}

type zebedeeRequestErrorData map[string]interface{}

func (e ErrInvalidZebedeeResponse) Error() string {
	return e.err.Error()
}

var _ error = ErrInvalidZebedeeResponse{}

// NewClient creates a new Zebedee Client
func NewClient(url string) ZebedeeClient {
	return ZebedeeClient{
		zebedeeURL: url,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (c ZebedeeClient) Get(path string) ([]byte, error) {
	return c.get(path)
}

func (c ZebedeeClient) GetLanding(path string) (StaticDatasetLandingPage, error) {
	b, err := c.get(path)
	if err != nil {
		return StaticDatasetLandingPage{}, err
	}

	dlp := new(models.DatasetLandingPage)
	if err = json.Unmarshal(b, &dlp); err != nil {
		return StaticDatasetLandingPage{}, err
	}

	related := [][]model.Related{
		dlp.RelatedDatasets,
		dlp.RelatedDocuments,
		dlp.RelatedMethodology,
		dlp.RelatedMethodologyArticle,
	}

	//Concurrently resolve any URIs where we need more data from another page
	var wg sync.WaitGroup
	sem := make(chan int, 10)

	for _, element := range related {
		for i, e := range element {
			sem <- 1
			wg.Add(1)
			go func(i int, e model.Related, element []model.Related) {
				defer func() {
					<-sem
					wg.Done()
				}()
				element[i].Title = c.getPageTitle(e.URI)
			}(i, e, element)
		}
	}
	wg.Wait()

	return c.populateStaticLandingPageModel(*dlp)
}

func (c ZebedeeClient) SetAccessToken(token string) {
	log.Debug("adding access token to client", log.Data{"token": token})
	c.accessToken = token
}

func (c ZebedeeClient) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.zebedeeURL+path, nil)
	if err != nil {
		return nil, err
	}

	if len(c.accessToken) > 0 {
		req.Header.Set("X-Florence-Token", c.accessToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidZebedeeResponse{errors.New("unexpected response from zebedee"), zebedeeRequestErrorData{
			"zebedeeURI":         req.URL.Path,
			"expectedStatusCode": 200,
			"actualStatusCode":   resp.StatusCode,
		}}
	}

	return ioutil.ReadAll(resp.Body)
}

func (c ZebedeeClient) getBreadcrumb(uri string) ([]model.TaxonomyNode, error) {
	b, err := c.get("/parents?uri=" + uri)
	if err != nil {
		return nil, err
	}

	var parentsJSON []models.TaxonomyNode
	if err = json.Unmarshal(b, &parentsJSON); err != nil {
		return nil, err
	}

	var parents []model.TaxonomyNode
	for _, value := range parentsJSON {
		parents = append(parents, model.TaxonomyNode{
			Title: value.Description.Title,
			URI:   value.URI,
		})
	}

	return parents, nil
}

func (c ZebedeeClient) populateStaticLandingPageModel(dlp models.DatasetLandingPage) (StaticDatasetLandingPage, error) {
	//Map Zebedee response data to new page model
	var sdlp StaticDatasetLandingPage
	sdlp.Type = dlp.Type
	sdlp.URI = dlp.URI
	sdlp.Metadata.Title = dlp.Description.Title
	sdlp.Metadata.Description = dlp.Description.Summary
	sdlp.DatasetLandingPage.Related.Datasets = dlp.RelatedDatasets
	sdlp.DatasetLandingPage.Related.Publications = dlp.RelatedDocuments
	sdlp.DatasetLandingPage.Related.Methodology = append(dlp.RelatedMethodology, dlp.RelatedMethodologyArticle...)
	sdlp.DatasetLandingPage.Related.Links = dlp.RelatedLinks
	sdlp.DatasetLandingPage.IsNationalStatistic = dlp.Description.NationalStatistic
	sdlp.DatasetLandingPage.IsTimeseries = dlp.Timeseries
	sdlp.ContactDetails.Email = dlp.Description.Contact.Email
	sdlp.ContactDetails.Telephone = dlp.Description.Contact.Telephone
	sdlp.ContactDetails.Name = dlp.Description.Contact.Name
	sdlp.DatasetLandingPage.ReleaseDate = dlp.Description.ReleaseDate
	sdlp.DatasetLandingPage.NextRelease = dlp.Description.NextRelease
	sdlp.DatasetLandingPage.Notes = dlp.Section.Markdown
	sdlp.FilterID = dlp.FilterID
	bc, err := c.getBreadcrumb(dlp.URI)
	if err != nil {
		return StaticDatasetLandingPage{}, err
	}
	sdlp.Page.Breadcrumb = bc
	sdlp.DatasetLandingPage.ParentPath = sdlp.Breadcrumb[len(sdlp.Breadcrumb)-1].Title

	for index, value := range dlp.Datasets {
		dataset := c.getDatasetDetails(value.URI)
		dataset.IsLast = index+1 == len(dlp.Datasets)

		sdlp.DatasetLandingPage.Datasets = append(sdlp.DatasetLandingPage.Datasets, dataset)
	}

	for _, value := range dlp.Alerts {
		switch value.Type {
		default:
			log.Debug("Unrecognised alert type", log.Data{"alert": value})
			fallthrough
		case "alert":
			sdlp.DatasetLandingPage.Notices = append(sdlp.DatasetLandingPage.Notices, datasetLandingPageStatic.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		case "correction":
			sdlp.DatasetLandingPage.Corrections = append(sdlp.DatasetLandingPage.Corrections, datasetLandingPageStatic.Message{
				Date:     value.Date,
				Markdown: value.Markdown,
			})
		}
	}

	return sdlp, nil
}

func (c ZebedeeClient) getDatasetDetails(uri string) datasetLandingPageStatic.Dataset {
	b, err := c.get("/data?uri=" + uri)
	if err != nil {
		log.Error(err, nil)
		return datasetLandingPageStatic.Dataset{URI: uri}
	}

	var d models.Dataset
	if err = json.Unmarshal(b, &d); err != nil {
		log.Error(err, nil)
		return datasetLandingPageStatic.Dataset{
			URI: uri,
		}
	}

	var dataset datasetLandingPageStatic.Dataset
	for _, value := range d.Downloads {
		dataset.Downloads = append(dataset.Downloads, datasetLandingPageStatic.Download{
			URI:       value.File,
			Extension: strings.TrimPrefix(filepath.Ext(value.File), "."),
			Size:      c.getFileSize(uri + "/" + value.File),
		})
	}
	for _, value := range d.SupplementaryFiles {
		dataset.SupplementaryFiles = append(dataset.SupplementaryFiles, datasetLandingPageStatic.SupplementaryFile{
			Title:     value.Title,
			URI:       value.File,
			Extension: strings.TrimPrefix(filepath.Ext(value.File), "."),
			Size:      c.getFileSize(uri + "/" + value.File),
		})
	}
	dataset.Title = d.Description.Edition
	if len(d.Versions) > 0 {
		dataset.HasVersions = true
	}

	dataset.URI = d.URI

	return dataset

}

func (c ZebedeeClient) getFileSize(uri string) string {
	b, err := c.get("/filesize?uri=" + uri)
	if err != nil {
		log.Error(err, nil)
		return ""
	}

	fs := struct {
		Size int `json:"fileSize"`
	}{}
	if err = json.Unmarshal(b, &fs); err != nil {
		log.Error(err, nil)
		return ""
	}

	return datasize.ByteSize(fs.Size).HumanReadable()
}

func (c ZebedeeClient) getPageTitle(uri string) string {
	b, err := c.get("/data?uri=" + uri + "&title")
	if err != nil {
		log.Error(err, nil)
		return ""
	}

	pt := struct {
		Title   string `json:"title"`
		Edition string `json:"edition"`
	}{}

	if err = json.Unmarshal(b, &pt); err != nil {
		log.Error(err, nil)
		return ""
	}

	if len(pt.Edition) > 0 && len(pt.Title) > 0 {
		return fmt.Sprintf("%s: %s", pt.Title, pt.Edition)
	}

	return pt.Title
}
