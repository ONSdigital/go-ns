package response

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"net/http/httptest"
	"net/http"
	"encoding/json"
)

type child struct {
	Name string `json:"value"`
}

type parent struct {
	Name  string `json:"name"`
	Child child `json:"child"`
}

type mockOnsJsonEncoder struct {
	invocationArgs  []interface{}
	mockedBehaviour func(w http.ResponseWriter, value interface{}, status int) error
}

func mockEncoderInit() *mockOnsJsonEncoder {
	impl := onsJSONEncoder{}
	return &mockOnsJsonEncoder{invocationArgs: make([]interface{}, 0), mockedBehaviour: impl.writeResponseJSON}
}

func (mock *mockOnsJsonEncoder) writeResponseJSON(w http.ResponseWriter, value interface{}, status int) error {
	mock.invocationArgs = append(mock.invocationArgs, value)
	return mock.mockedBehaviour(w, value, status)
}

func TestWriteJSON(t *testing.T) {

	Convey("Should Write value to response as JSON, set the expected http response status code & set the content " +
		"type header with the expected value", t, func() {
		rec := httptest.NewRecorder()
		mock := mockEncoderInit()
		jsonResponseEncoder = mock

		expected := parent{Name: "Hello World!", Child: child{Name: "Bob!"}}

		WriteJSON(rec, expected, http.StatusOK)

		var actual parent
		json.Unmarshal(rec.Body.Bytes(), &actual)

		So(actual, ShouldResemble, expected)
		So(rec.Code, ShouldEqual, http.StatusOK)
		So(rec.Header().Get(contentTypeHeader), ShouldEqual, contentTypeJSON)
		So(len(mock.invocationArgs), ShouldEqual, 1)
		So(mock.invocationArgs[0], ShouldResemble, expected)
	})

	Convey("Should return internal server error http status if the writer cannot successfully " +
		"write to the response", t, func() {
		mock := mockEncoderInit()
		jsonResponseEncoder = mock

		rec := httptest.NewRecorder()

		// a function cannot be marshaled into JSON is an will cause he JSON encoder to throw an error.
		arg := func() string {
			return "HelloWorld"
		}

		WriteJSON(rec, arg, http.StatusOK)

		So(rec.Code, ShouldEqual, http.StatusInternalServerError)
		So(rec.Header().Get(contentTypeHeader), ShouldEqual, contentTypeJSON)
		So(len(mock.invocationArgs), ShouldEqual, 1)
	})
}