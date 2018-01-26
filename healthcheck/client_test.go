package healthcheck

//go:generate moq -out mock_healthcheck/mock_httpclient.go -pkg mock_healthcheck . HttpClient http.Response

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ONSdigital/go-ns/healthcheck/mock_healthcheck"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHealthcheckClient(t *testing.T) {

	service := "myService"

	Convey("Healthcheck returns nil when no error", t, func() {
		mock := &mock_healthcheck.HttpClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				return &http.Response{StatusCode:http.StatusOK}, nil
			},
		}

		client := healthcheckClient{Client:mock, Url:"http://foo/bar", Service: service}

		s, e := client.Healthcheck()
		So(s, ShouldEqual, service)
		So(e, ShouldBeNil)
	})

	Convey("Healthcheck returns error when status != 200", t, func() {
		mock := &mock_healthcheck.HttpClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				return &http.Response{StatusCode:http.StatusInternalServerError}, nil
			},
		}

		client := healthcheckClient{Client:mock, Url:"http://foo/bar", Service: service}

		s, e := client.Healthcheck()
		So(s, ShouldEqual, service)
		So(e, ShouldNotBeNil)
	})

	Convey("Healthcheck returns error when Get returns error", t, func() {
		err := errors.New("This is an error")
		mock := &mock_healthcheck.HttpClientMock{
			GetFunc: func(url string) (*http.Response, error) {
				return &http.Response{StatusCode:http.StatusInternalServerError}, err
			},
		}

		client := healthcheckClient{Client:mock, Url:"http://foo/bar", Service: service}

		s, e := client.Healthcheck()
		So(s, ShouldEqual, service)
		So(e, ShouldEqual, err)
	})

}
