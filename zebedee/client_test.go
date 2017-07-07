package zebedee

import (
	"net/http"
	"os"
	"testing"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitClient(t *testing.T) {
	go mockZebedeeServer()

	cli := NewClient("http://localhost:3050")

	Convey("test Get()", t, func() {
		Convey("test get sucessfully returns response from zebedee", func() {
			b, err := cli.Get("/data?uri=foo")
			So(err, ShouldBeNil)

			So(string(b), ShouldEqual, `{}`)
		})

		Convey("test error returned if requesting invalid zebedee url", func() {
			b, err := cli.Get("/invalid")
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
			So(err.Error(), ShouldEqual, "unexpected response from zebedee")
			So(b, ShouldBeNil)
		})
	})

	Convey("test GetLanding", t, func() {
		Convey("test getLanding sucessfully returns a landing model", func() {
			m, err := cli.GetLanding("/data?uri=labor")
			So(err, ShouldBeNil)
			So(m, ShouldNotBeEmpty)
			So(m.Page.Type, ShouldEqual, "dataset_landing_page")
		})

		Convey("test error returned if requesting invalid zebedee url", func() {
			_, err := cli.GetLanding("/invalid")
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
			So(err.Error(), ShouldEqual, "unexpected response from zebedee")
		})
	})

	Convey("test get dataset details", t, func() {
		d := cli.getDatasetDetails("12345")
		So(d.URI, ShouldEqual, "www.google.com")
		So(d.SupplementaryFiles[0].Title, ShouldEqual, "helloworld")
	})

	Convey("test getFileSize returns human readable filesize", t, func() {
		fs := cli.getFileSize("filesize")
		So(fs, ShouldEqual, "5.0 MB")
	})

	Convey("test getPageTitle returns a correctly formatted page title", t, func() {
		t := cli.getPageTitle("pageTitle")
		So(t, ShouldEqual, "baby-names: 2017")
	})
}

func mockZebedeeServer() {
	r := mux.NewRouter()

	r.Path("/data").HandlerFunc(data)
	r.Path("/parents").HandlerFunc(parents)
	r.Path("/filesize").HandlerFunc(filesize)

	s := server.New(":3050", r)

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		os.Exit(2)
	}
}

func data(w http.ResponseWriter, req *http.Request) {
	uri := req.URL.Query().Get("uri")

	switch uri {
	case "foo":
		w.Write([]byte(`{}`))
	case "labor":
		w.Write([]byte(`{"downloads":[{"title":"Latest","file":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02/labd02jul2015_tcm77-408195.xls"}],"section":{"markdown":""},"relatedDatasets":[{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputeslabd01"},{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/stoppagesofworklabd03"}],"relatedDocuments":[{"uri":"/employmentandlabourmarket/peopleinwork/employmentandemployeetypes/bulletins/uklabourmarket/2015-07-15"}],"relatedMethodology":[],"type":"dataset_landing_page","uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02","description":{"title":"Labour disputes by sector: LABD02","summary":"Labour disputes by sector.","keywords":["strike"],"metaDescription":"Labour disputes by sector.","nationalStatistic":true,"contact":{"email":"richard.clegg@ons.gsi.gov.uk\n","name":"Richard Clegg\n","telephone":"+44 (0)1633 455400 \n"},"releaseDate":"2015-07-14T23:00:00.000Z","nextRelease":"12 August 2015","datasetId":"","unit":"","preUnit":"","source":""}}`))
	case "12345":
		w.Write([]byte(`{"type":"dataset","uri":"www.google.com","downloads":[{"file":"test.txt"}],"supplementaryFiles":[{"title":"helloworld","file":"helloworld.txt"}],"versions":[{"uri":"www.google.com"}]}`))
	case "pageTitle":
		w.Write([]byte(`{"title":"baby-names","edition":"2017"}`))
	}

}

func parents(w http.ResponseWriter, req *http.Request) {
	uri := req.URL.Query().Get("uri")

	switch uri {
	case "/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02":
		w.Write([]byte(`[{"uri":"/","description":{"title":"Home"},"type":"home_page"},{"uri":"/employmentandlabourmarket","description":{"title":"Employment and labour market"},"type":"taxonomy_landing_page"},{"uri":"/employmentandlabourmarket/peopleinwork","description":{"title":"People in work"},"type":"taxonomy_landing_page"},{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions","description":{"title":"Workplace disputes and working conditions"},"type":"product_page"}]`))
	}
}

func filesize(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(`{"fileSize":5242880}`))
}
