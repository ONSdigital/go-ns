package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rhttp"
	"github.com/ONSdigital/go-ns/zebedee/data"
)

// ZebedeeClient represents a zebedee client
type ZebedeeClient struct {
	zebedeeURL string
	client     *rhttp.Client
}

// ErrInvalidZebedeeResponse is returned when zebedee does not respond
// with a valid status
type ErrInvalidZebedeeResponse struct {
	actualCode int
	uri        string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidZebedeeResponse) Error() string {
	return fmt.Sprintf("invalid response from zebedee - should be 2.x.x or 3.x.x, got: %d, path: %s",
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidZebedeeResponse{}

// NewZebedeeClient creates a new Zebedee Client, set ZEBEDEE_REQUEST_TIMEOUT_SECOND
// environment variable to modify default client timeout as zebedee can often be slow
// to respond
func NewZebedeeClient(url string) *ZebedeeClient {
	timeout, err := strconv.Atoi(os.Getenv("ZEBEDEE_REQUEST_TIMEOUT_SECONDS"))
	if timeout == 0 || err != nil {
		timeout = 5
	}
	cli := rhttp.DefaultClient
	cli.HTTPClient.Timeout = time.Duration(timeout) * time.Second

	return &ZebedeeClient{
		zebedeeURL: url,
		client:     cli,
	}
}

// Get returns a response for the requested uri in zebedee
func (c *ZebedeeClient) Get(ctx context.Context, path string) ([]byte, error) {
	return c.get(ctx, path)
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *ZebedeeClient) Healthcheck() (string, error) {
	resp, err := c.client.Get(c.zebedeeURL + "/healthcheck")
	if err != nil {
		return "zebedee", err
	}

	if resp.StatusCode != http.StatusOK {
		return "zebedee", ErrInvalidZebedeeResponse{resp.StatusCode, "/healthcheck"}
	}

	return "", nil
}

// GetDatasetLandingPage returns a DatasetLandingPage populated with data from a zebedee response. If an error
// is returned there is a chance that a partly completed DatasetLandingPage is returned
func (c *ZebedeeClient) GetDatasetLandingPage(ctx context.Context, path string) (data.DatasetLandingPage, error) {
	reqURL := c.createRequestURL(ctx, "/data?uri="+path)
	b, err := c.get(ctx, reqURL)
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
				t, _ := c.GetPageTitle(ctx, e.URI)
				element[i].Title = t.Title
			}(i, e, element)
		}
	}
	wg.Wait()

	return dlp, nil
}

func (c *ZebedeeClient) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.zebedeeURL+path, nil)
	if err != nil {
		return nil, err
	}

	if ctx.Value(common.FlorenceIdentityKey) != nil {
		accessToken, ok := ctx.Value(common.FlorenceIdentityKey).(string)
		if !ok {
			log.ErrorCtx(ctx, errors.New("error casting access token cookie to string"), nil)
		}
		req.Header.Set(common.FlorenceHeaderKey, accessToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		io.Copy(ioutil.Discard, resp.Body)
		return nil, ErrInvalidZebedeeResponse{resp.StatusCode, req.URL.Path}
	}

	return ioutil.ReadAll(resp.Body)
}

// GetBreadcrumb returns a Breadcrumb
func (c *ZebedeeClient) GetBreadcrumb(ctx context.Context, uri string) ([]data.Breadcrumb, error) {
	b, err := c.get(ctx, "/parents?uri="+uri)
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
func (c *ZebedeeClient) GetDataset(ctx context.Context, uri string) (data.Dataset, error) {
	b, err := c.get(ctx, "/data?uri="+uri)
	if err != nil {
		return data.Dataset{}, err
	}

	var d data.Dataset
	if err = json.Unmarshal(b, &d); err != nil {
		return d, err
	}

	downloads := make([]data.Download, 0)

	for _, v := range d.Downloads {
		fs, err := c.GetFileSize(ctx, uri+"/"+v.File)
		if err != nil {
			return d, err
		}

		downloads = append(downloads, data.Download{
			File: v.File,
			Size: strconv.Itoa(fs.Size),
		})
	}

	d.Downloads = downloads

	supplementaryFiles := make([]data.SupplementaryFile, 0)
	for _, v := range d.SupplementaryFiles {
		fs, err := c.GetFileSize(ctx, uri+"/"+v.File)
		if err != nil {
			return d, err
		}

		supplementaryFiles = append(supplementaryFiles, data.SupplementaryFile{
			File:  v.File,
			Title: v.Title,
			Size:  strconv.Itoa(fs.Size),
		})
	}

	d.SupplementaryFiles = supplementaryFiles

	return d, nil
}

// GetFileSize retrieves a given filesize from zebedee
func (c *ZebedeeClient) GetFileSize(ctx context.Context, uri string) (data.FileSize, error) {
	b, err := c.get(ctx, "/filesize?uri="+uri)
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
func (c *ZebedeeClient) GetPageTitle(ctx context.Context, uri string) (data.PageTitle, error) {
	b, err := c.get(ctx, "/data?uri="+uri+"&title")
	if err != nil {
		return data.PageTitle{}, err
	}

	var pt data.PageTitle
	if err = json.Unmarshal(b, &pt); err != nil {
		return pt, err
	}

	return pt, nil
}

func (c *ZebedeeClient) createRequestURL(ctx context.Context, path string) string {
	var url string
	if ctx.Value(common.CollectionIDHeaderKey) != nil {
		collectionID, ok := ctx.Value(common.CollectionIDHeaderKey).(string)
		if !ok {
			log.ErrorCtx(ctx, errors.New("error casting collection ID cookie to string"), nil)
		}
		url = "/data/" + collectionID
	}
	url = url + path
	return url
}
