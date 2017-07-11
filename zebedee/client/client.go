package client

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ONSdigital/go-ns/zebedee/data"
)

// ZebedeeClient represents a zebedee client
type ZebedeeClient struct {
	zebedeeURL  string
	client      *http.Client
	accessToken string
}

// ErrInvalidZebedeeResponse is returned when zebedee does not respond
// with a valid status
type ErrInvalidZebedeeResponse struct {
	actualCode   int
	expectedCode int
	uri          string
}

func (e ErrInvalidZebedeeResponse) Error() string {
	return fmt.Sprintf("invalid response from zebedee - expected: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidZebedeeResponse{}

// NewZebedeeClient creates a new Zebedee Client
func NewZebedeeClient(url string) ZebedeeClient {
	return ZebedeeClient{
		zebedeeURL: url,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

// Get returns a response for the requested uri in zebedee
func (c ZebedeeClient) Get(path string) ([]byte, error) {
	return c.get(path)
}

// GetDatasetLandingPage returns a DatasetLandingPage populated with data from a zebedee response. If an error
// is returned there is a chance that a partly completed DatasetLandingPage is returned
func (c ZebedeeClient) GetDatasetLandingPage(path string) (data.DatasetLandingPage, error) {
	b, err := c.get(path)
	if err != nil {
		return data.DatasetLandingPage{}, err
	}

	var dlp data.DatasetLandingPage
	if err = json.Unmarshal(b, &dlp); err != nil {
		return dlp, err
	}

	related := [][]data.Related{
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
			go func(i int, e data.Related, element []data.Related) {
				defer func() {
					<-sem
					wg.Done()
				}()
				t, _ := c.GetPageTitle(e.URI)
				element[i].Title = t.Title
			}(i, e, element)
		}
	}
	wg.Wait()

	return dlp, nil
}

// SetAccessToken adds an access token to the client to authenticate with zebedee
func (c ZebedeeClient) SetAccessToken(token string) {
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
		io.Copy(ioutil.Discard, resp.Body)
		return nil, ErrInvalidZebedeeResponse{resp.StatusCode, http.StatusOK, req.URL.Path}
	}

	return ioutil.ReadAll(resp.Body)
}

// GetBreadcrumb returns a Breadcrumb
func (c ZebedeeClient) GetBreadcrumb(uri string) ([]data.Breadcrumb, error) {
	b, err := c.get("/parents?uri=" + uri)
	if err != nil {
		return nil, err
	}

	var parentsJSON []data.Breadcrumb
	if err = json.Unmarshal(b, &parentsJSON); err != nil {
		return nil, err
	}

	return parentsJSON, nil
}

// GetDataset returns details about a dataset from zebedee
func (c ZebedeeClient) GetDataset(uri string) (data.Dataset, error) {
	b, err := c.get("/data?uri=" + uri)
	if err != nil {
		return data.Dataset{}, err
	}

	var d data.Dataset
	if err = json.Unmarshal(b, &d); err != nil {
		return d, err
	}

	for _, v := range d.Downloads {
		fs, err := c.GetFileSize(uri + "/" + v.File)
		if err != nil {
			return d, err
		}

		v.Size = fs.Size
	}

	for _, v := range d.SupplementaryFiles {
		fs, err := c.GetFileSize(uri + "/" + v.File)
		if err != nil {
			return d, err
		}

		v.Size = fs.Size
	}

	return d, nil
}

// GetFileSize retrieves a given filesize from zebedee
func (c ZebedeeClient) GetFileSize(uri string) (data.FileSize, error) {
	b, err := c.get("/filesize?uri=" + uri)
	if err != nil {
		return data.FileSize{}, err
	}

	var fs data.FileSize
	if err = json.Unmarshal(b, &fs); err != nil {
		return fs, err
	}

	return fs, nil
}

// GetPageTitle retrieves a page title from zebedee
func (c ZebedeeClient) GetPageTitle(uri string) (data.PageTitle, error) {
	b, err := c.get("/data?uri=" + uri + "&title")
	if err != nil {
		return data.PageTitle{}, err
	}

	var pt data.PageTitle
	if err = json.Unmarshal(b, &pt); err != nil {
		return pt, err
	}

	return pt, nil
}
