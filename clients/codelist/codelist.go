package codelist

import (
	"context"
	"encoding/json"
	"fmt"
	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"
	"io"
	"io/ioutil"
	"net/http"
)

//go:generate moq -out codelisttest/genrated_client_mock.go -pkg codelisttest . Storer

const service = "code-list-api"

var _ error = ErrInvalidCodelistAPIResponse{}

// Client is a codelist api client which can be used to make requests to the server
type Client struct {
	cli rchttp.Clienter
	url string
}

// ErrInvalidCodelistAPIResponse is returned when the codelist api does not respond
// with a valid status
type ErrInvalidCodelistAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidCodelistAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from codelist api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

// Code returns the status code received from code list api if an error is returned
func (e ErrInvalidCodelistAPIResponse) Code() int {
	return e.actualCode
}

// New creates a new instance of Client with a given filter api url
func New(codelistAPIURL string) *Client {
	return &Client{
		cli: rchttp.NewClient(),
		url: codelistAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() (string, error) {
	resp, err := c.cli.Get(context.Background(), c.url+"/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return service, nil
}

func (c *Client) doWithAuthToken(ctx context.Context, method string, uri string, authToken string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	common.AddServiceTokenHeader(req, authToken)
	return c.cli.Do(ctx, req)
}

// GetValues returns dimension values from the codelist api
func (c *Client) GetValues(ctx context.Context, authToken string, id string) (DimensionValues, error) {
	var vals DimensionValues
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.url, id)

	clientlog.Do(context.Background(), "retrieving codes from codelist", service, uri)

	resp, err := c.doWithAuthToken(ctx, "GET", uri, authToken, nil)
	if err != nil {
		return vals, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return vals, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return vals, err
	}

	err = json.Unmarshal(b, &vals)
	return vals, err
}

// GetIDNameMap returns dimension values in the form of an id name map
func (c *Client) GetIDNameMap(ctx context.Context, id string, authToken string) (map[string]string, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.url, id)

	resp, err := c.doWithAuthToken(ctx, "GET", uri, authToken, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Event(ctx, "error closing codelistClient.GetIDNameMap response body", log.Error(err))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var vals DimensionValues
	if err = json.Unmarshal(body, &vals); err != nil {
		return nil, err
	}

	idNames := make(map[string]string)
	for _, val := range vals.Items {
		idNames[val.ID] = val.Label
	}

	return idNames, nil
}

// GetIDNameMap returns dimension values in the form of an id name map
/*func (c *Client) GetIDNameMap(id string) (map[string]string, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.url, id)
	resp, err := c.cli.Get(context.Background(), uri)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var vals DimensionValues
	if err = json.Unmarshal(b, &vals); err != nil {
		return nil, err
	}

	idNames := make(map[string]string)
	for _, val := range vals.Items {
		idNames[val.ID] = val.Label
	}

	return idNames, nil
}*/

//GetGeographyCodeLists returns the geography codelists
func (c *Client) GetGeographyCodeLists() (results CodeListResults, err error) {
	uri := fmt.Sprintf("%s/code-lists?type=geography", c.url)
	resp, err := c.cli.Get(context.Background(), uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &results)
	if err != nil {
		return
	}
	return results, nil
}

//GetCodeListEditions returns the editions for a codelist
func (c *Client) GetCodeListEditions(codeListID string) (editions EditionsListResults, err error) {
	url := fmt.Sprintf("%s/code-lists/%s/editions", c.url, codeListID)
	resp, err := c.cli.Get(context.Background(), url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &editions)
	if err != nil {
		return
	}
	return editions, nil
}

//GetCodes returns the codes for a specific edition of a code list
func (c *Client) GetCodes(codeListID string, edition string) (codes CodesResults, err error) {
	url := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes", c.url, codeListID, edition)
	resp, err := c.cli.Get(context.Background(), url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &codes)
	if err != nil {
		return
	}
	return codes, nil
}

// GetCodeByID returns informtion about a code
func (c *Client) GetCodeByID(codeListID string, edition string, codeID string) (code CodeResult, err error) {
	url := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes/%s", c.url, codeListID, edition, codeID)
	resp, err := c.cli.Get(context.Background(), url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &code)
	if err != nil {
		return
	}
	return code, nil
}

func (c *Client) GetDatasetsByCode(codeListID string, edition string, codeID string) (datasets DatasetsResult, err error) {
	url := fmt.Sprintf("%s/code-lists/%s/editions/%s/codes/%s/datasets", c.url, codeListID, edition, codeID)
	resp, err := c.cli.Get(context.Background(), url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &datasets)
	if err != nil {
		return
	}
	return datasets, nil
}
