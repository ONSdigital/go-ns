package clientlog

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/ONSdigital/go-ns/log"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitClientlog(t *testing.T) {
	ctx := context.Background()

	Convey("test ouput produced by client log package with no log.Data", t, func() {
		log.Namespace = "dp-frontend-service"

		output := captureOutput(func() {
			Do(ctx, "retrieving datasets", "dp-backend-api", "http://localhost:22000/datasets")
		})

		logContents := make(map[string]interface{})

		err := json.Unmarshal([]byte(output), &logContents)
		So(err, ShouldBeNil)

		So(logContents["created"].(string), ShouldNotBeEmpty)
		So(logContents["event"].(string), ShouldEqual, "info")
		So(logContents["namespace"].(string), ShouldEqual, "dp-frontend-service")

		logData := logContents["data"].(map[string]interface{})
		So(logData, ShouldNotBeEmpty)
		So(logData["action"].(string), ShouldEqual, "retrieving datasets")
		So(logData["message"].(string), ShouldEqual, "Making request to service: dp-backend-api")
		So(logData["method"].(string), ShouldEqual, "GET")
		So(logData["uri"].(string), ShouldEqual, "http://localhost:22000/datasets")
	})

	Convey("test output produced by client log package with log.Data", t, func() {
		log.Namespace = "dp-frontend-service"

		output := captureOutput(func() {
			Do(ctx, "retrieving datasets", "dp-backend-api", "http://localhost:22000/datasets", log.Data{
				"method": "DELETE",
				"value":  "abcdefg",
			})
		})

		logContents := make(map[string]interface{})

		err := json.Unmarshal([]byte(output), &logContents)
		So(err, ShouldBeNil)

		So(logContents["created"].(string), ShouldNotBeEmpty)
		So(logContents["event"].(string), ShouldEqual, "info")
		So(logContents["namespace"].(string), ShouldEqual, "dp-frontend-service")

		logData := logContents["data"].(map[string]interface{})
		So(logData, ShouldNotBeEmpty)
		So(logData["action"].(string), ShouldEqual, "retrieving datasets")
		So(logData["message"].(string), ShouldEqual, "Making request to service: dp-backend-api")
		So(logData["method"].(string), ShouldEqual, "DELETE")
		So(logData["value"].(string), ShouldEqual, "abcdefg")
		So(logData["uri"].(string), ShouldEqual, "http://localhost:22000/datasets")
	})
}

func captureOutput(f func()) string {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan string)

	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	w.Close()
	os.Stdout = stdout
	out := <-outC
	return out
}
