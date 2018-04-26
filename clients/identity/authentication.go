package identity

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/rchttp"
)

type IdentityClient common.APIClient

type IdentityClienter interface {
	CheckRequest(req *http.Request) (context.Context, int, error)
}

// NewAPIClient returns an IdentityClient
func NewAPIClient(cli common.RCHTTPClienter, url string) (api *IdentityClient) {
	return &IdentityClient{
		HTTPClient: cli,
		BaseURL:    url,
	}
}

// CheckRequest calls the AuthAPI to check florenceToken or authToken
func (api IdentityClient) CheckRequest(req *http.Request) (context.Context, int, error) {

	log.DebugR(req, "CheckRequest called", nil)

	ctx := req.Context()

	florenceToken := req.Header.Get(common.FlorenceHeaderKey)
	authToken := req.Header.Get(common.AuthHeaderKey)

	isUserReq := len(florenceToken) > 0
	isServiceReq := len(authToken) > 0

	// if neither user nor service request, return unchanged ctx
	if !isUserReq && !isServiceReq {
		return ctx, http.StatusUnauthorized, nil
	}

	url := api.BaseURL + "/identity"
	logData := log.Data{"is_user_request": isUserReq, "is_service_request": isServiceReq, "url": url}
	log.DebugR(req, "calling AuthAPI to authenticate", logData)

	outboundAuthReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.ErrorR(req, err, logData)
		return nil, 0, err
	}

	if isUserReq {
		outboundAuthReq.Header.Set(common.FlorenceHeaderKey, florenceToken)
	} else {
		outboundAuthReq.Header.Set(common.AuthHeaderKey, authToken)
	}

	if api.HTTPClient == nil {
		api.Lock.Lock()
		api.HTTPClient = rchttp.NewClient()
		api.Lock.Unlock()
	}

	resp, err := api.HTTPClient.Do(ctx, outboundAuthReq)
	if err != nil {
		log.ErrorR(req, err, logData)
		return nil, 0, err
	}

	// Check to see if the user is authorised
	if resp.StatusCode != http.StatusOK {
		logData["status"] = resp.StatusCode
		log.DebugR(req, "unexpected status code returned from AuthAPI", logData)
		return nil, resp.StatusCode, nil
	}

	identityResp, err := unmarshalIdentityResponse(resp)
	if err != nil {
		log.ErrorR(req, err, logData)
		return nil, 0, err
	}

	var userIdentity string
	if isUserReq {
		userIdentity = identityResp.Identifier
	} else {
		userIdentity = req.Header.Get(common.UserHeaderKey)
	}

	logData["user_identity"] = userIdentity
	logData["caller_identity"] = identityResp.Identifier
	log.DebugR(req, "identity retrieved, setting context values", logData)

	ctx = context.WithValue(ctx, common.UserIdentityKey, userIdentity)
	ctx = context.WithValue(ctx, common.CallerIdentityKey, identityResp.Identifier)

	return ctx, http.StatusOK, nil
}

// unmarshalIdentityResponse converts a resp.Body (JSON) into an IdentityResponse
func unmarshalIdentityResponse(resp *http.Response) (identityResp *common.IdentityResponse, err error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.ErrorC("error closing response body", err, nil)
		}
	}()

	err = json.Unmarshal(b, &identityResp)
	return
}
