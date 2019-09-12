package renderer

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	rchttp "github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/clients/clientlog"
	"github.com/ONSdigital/go-ns/log"
)

const service = "renderer"

// ErrInvalidRendererResponse is returned when the renderer service does not respond
// with a status 200
type ErrInvalidRendererResponse struct {
	responseCode int
}

// Error should be called by the user to print out the stringified version of the error
func (e ErrInvalidRendererResponse) Error() string {
	return fmt.Sprintf("invalid response from renderer service - status %d", e.responseCode)
}

// Code returns the status code received from renderer if an error is returned
func (e ErrInvalidRendererResponse) Code() int {
	return e.responseCode
}

// Renderer represents a renderer client to interact with the dp-frontend-renderer
type Renderer struct {
	client rchttp.Clienter
	url    string
}

// New creates an instance of renderer with a default client
func New(url string) *Renderer {
	return &Renderer{
		client: rchttp.NewClient(),
		url:    url,
	}
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.ErrorCtx(ctx, err, log.Data{"Message": "error closing http response body"})
	}
}

// Healthcheck calls the healthcheck endpoint on the renderer and returns any errors
func (r *Renderer) Healthcheck() (string, error) {
	resp, err := r.client.Get(context.Background(), r.url+"/healthcheck")
	if err != nil {
		return service, err
	}

	if resp.StatusCode != http.StatusOK {
		return service, ErrInvalidRendererResponse{resp.StatusCode}
	}

	return service, nil
}

// Do sends a request to the renderer service to render a given template
func (r *Renderer) Do(path string, b []byte) ([]byte, error) {
	// Renderer required JSON to be sent so if byte array is empty, set it to be
	// empty json
	if b == nil {
		b = []byte(`{}`)
	}

	uri := r.url + "/" + path
	ctx := context.Background()

	clientlog.Do(ctx, fmt.Sprintf("rendering template: %s", path), service, uri, log.Data{
		"method": "POST",
	})

	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponseBody(ctx, resp)

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidRendererResponse{resp.StatusCode}
	}

	return ioutil.ReadAll(resp.Body)
}
