package healthcheck_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ONSdigital/go-ns/healthcheck"
	"github.com/ONSdigital/dp-rchttp"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHealthcheckClientWithoutError(t *testing.T) {

	service := "myService"
	url := "http://foo/bar"

	Convey("Given a healthcheck client with mocked HttpClient", t, func() {
		mock := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
		}

		client := healthcheck.NewClient(service, url, mock)

		Convey("When healthcheck is invoked", func() {

			s, err := client.Healthcheck()

			Convey("Healthcheck invokes HttpClient correctly and returns nil", func() {

				So(s, ShouldEqual, service)
				So(err, ShouldBeNil)
				So(len(mock.GetCalls()), ShouldEqual, 1)
				So(mock.GetCalls()[0].URL, ShouldEqual, url)
			})
		})
	})
}

func TestHealthcheckClientReportsError(t *testing.T) {

	service := "myService"

	Convey("Given a healthcheck client with mocked HttpClient returning a 500 error", t, func() {
		mock := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError}, nil
			},
		}

		client := healthcheck.NewClient(service, "http://foo/bar", mock)

		Convey("When healthcheck is invoked", func() {

			s, err := client.Healthcheck()

			Convey("Healthcheck returns an error", func() {

				So(s, ShouldEqual, service)
				So(err, ShouldNotBeNil)
			})
		})
	})

}

func TestHealthcheckClientReturnsError(t *testing.T) {

	service := "myService"

	Convey("Given a healthcheck client with mocked HttpClient returning a connection error", t, func() {
		mockErr := errors.New("This is an error")
		mock := &rchttp.ClienterMock{
			GetFunc: func(ctx context.Context, url string) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusInternalServerError}, mockErr
			},
		}

		client := healthcheck.NewClient(service, "http://foo/bar", mock)

		Convey("When healthcheck is invoked", func() {

			s, err := client.Healthcheck()

			Convey("Healthcheck returns the error error", func() {

				So(s, ShouldEqual, service)
				So(err, ShouldEqual, mockErr)
			})
		})
	})

}
