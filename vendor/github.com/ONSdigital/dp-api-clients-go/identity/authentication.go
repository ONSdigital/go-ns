package identity

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go"
	"github.com/ONSdigital/dp-api-clients-go/headers"
	"github.com/ONSdigital/dp-rchttp"
	"github.com/ONSdigital/go-ns/common"
	"github.com/ONSdigital/log.go/log"

	"github.com/pkg/errors"
)

var errUnableToIdentifyRequest = errors.New("unable to determine the user or service making the request")

type tokenObject struct {
	numberOfParts int
	hasPrefix     bool
	tokenPart     string
}

// Client is an alias to a generic/common api client structure
type Client clients.APIClient

// Clienter provides an interface to checking identity of incoming request
type Clienter interface {
	CheckRequest(req *http.Request) (context.Context, int, authFailure, error)
}

// NewAPIClient returns a Client
func NewAPIClient(cli rchttp.Clienter, url string) (api *Client) {
	return &Client{
		HTTPClient: cli,
		BaseURL:    url,
	}
}

// authFailure is an alias to an error type, this represents the failure to
// authenticate request over a generic error from a http or marshalling error
type authFailure error

// CheckRequest calls the AuthAPI to check florenceToken or serviceAuthToken
func (api Client) CheckRequest(req *http.Request, florenceToken, serviceAuthToken string) (context.Context, int, authFailure, error) {
	ctx := req.Context()

	isUserReq := len(florenceToken) > 0
	isServiceReq := len(serviceAuthToken) > 0

	// if neither user nor service request, return unchanged ctx
	if !isUserReq && !isServiceReq {
		return ctx, http.StatusUnauthorized, errors.WithMessage(errUnableToIdentifyRequest, "no headers set on request"), nil
	}

	url := api.BaseURL + "/identity"

	logData := log.Data{
		"is_user_request":    isUserReq,
		"is_service_request": isServiceReq,
		"url":                url,
	}
	splitTokens(florenceToken, serviceAuthToken, logData)

	log.Event(ctx, "calling AuthAPI to authenticate caller identity", logData)

	outboundAuthReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Event(ctx, "error creating AuthAPI identity http request", logData, log.Error(err))
		return ctx, http.StatusInternalServerError, nil, err
	}

	if isUserReq {
		headers.SetUserAuthToken(outboundAuthReq, florenceToken)
	} else {
		headers.SetServiceAuthToken(outboundAuthReq, serviceAuthToken)
	}

	if api.HTTPClient == nil {
		api.Lock.Lock()
		api.HTTPClient = rchttp.NewClient()
		api.Lock.Unlock()
	}

	resp, err := api.HTTPClient.Do(ctx, outboundAuthReq)
	if err != nil {
		log.Event(ctx, "HTTPClient.Do returned error making AuthAPI identity request", logData, log.Error(err))
		return ctx, http.StatusInternalServerError, nil, err
	}

	defer closeResponse(ctx, resp, logData)

	// Check to see if the user is authorised
	if resp.StatusCode != http.StatusOK {
		return ctx, resp.StatusCode, errors.WithMessage(errUnableToIdentifyRequest, "unexpected status code returned from AuthAPI"), nil
	}

	identityResp, err := unmarshalIdentityResponse(resp)
	if err != nil {
		log.Event(ctx, "error unmarshaling AuthAPI identity response body", logData, log.Error(err))
		return ctx, http.StatusInternalServerError, nil, err
	}

	var userIdentity string
	if isUserReq {
		userIdentity = identityResp.Identifier
	} else {
		userIdentity, _  = headers.GetUserIdentity(req)
	}

	logData["user_identity"] = userIdentity
	logData["caller_identity"] = identityResp.Identifier
	log.Event(ctx, "caller identity retrieved setting context values", logData)

	ctx = context.WithValue(ctx, common.UserIdentityKey, userIdentity)
	ctx = context.WithValue(ctx, common.CallerIdentityKey, identityResp.Identifier)

	return ctx, http.StatusOK, nil, nil
}

func splitTokens(florenceToken, authToken string, logData log.Data) {
	if len(florenceToken) > 0 {
		logData["florence_token"] = splitToken(florenceToken)
	}
	if len(authToken) > 0 {
		logData["auth_token"] = splitToken(authToken)
	}
}

func splitToken(token string) (tokenObj tokenObject) {
	splitToken := strings.Split(token, " ")
	tokenObj.numberOfParts = len(splitToken)
	tokenObj.hasPrefix = strings.HasPrefix(token, common.BearerPrefix)

	// sample last 6 chars (or half, if smaller) of last token part
	lastTokenPart := len(splitToken) - 1
	tokenSampleStart := len(splitToken[lastTokenPart]) - 6
	if tokenSampleStart < 1 {
		tokenSampleStart = len(splitToken[lastTokenPart]) / 2
	}
	tokenObj.tokenPart = splitToken[lastTokenPart][tokenSampleStart:]

	return tokenObj
}

// unmarshalIdentityResponse converts a resp.Body (JSON) into an IdentityResponse
func unmarshalIdentityResponse(resp *http.Response) (identityResp *common.IdentityResponse, err error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &identityResp)
	return
}

func closeResponse(ctx context.Context, resp *http.Response, data log.Data) {
	if resp == nil || resp.Body == nil {
		return
	}

	if errClose := resp.Body.Close(); errClose != nil {
		log.Event(ctx, "error closing response body", log.Error(errClose), data)
	}
}
