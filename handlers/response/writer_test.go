package response

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"encoding/json"
)

type child struct {
	Name string `json:"value"`
}

type parent struct {
	Name  string `json:"name"`
	Child child  `json:"child"`
}

type mockOnsJSONEncoder struct {
	invocationArgs  []interface{}
	mockedBehaviour func(w http.ResponseWriter, value interface{}, status int) error
}

func mockEncoderInit() *mockOnsJSONEncoder {
	impl := onsJSONEncoder{}
	return &mockOnsJSONEncoder{invocationArgs: make([]interface{}, 0), mockedBehaviour: impl.writeResponseJSON}
}

func (mock *mockOnsJSONEncoder) writeResponseJSON(w http.ResponseWriter, value interface{}, status int) error {
	mock.invocationArgs = append(mock.invocationArgs, value)
	return mock.mockedBehaviour(w, value, status)
}

func TestWriteJSON(t *testing.T) {
	var input parent
	var statusCode int
	var rec *httptest.ResponseRecorder
	mock := mockEncoderInit()
	jsonResponseEncoder = mock

	Convey("Given a valid responseWriter, response value and status code", t, func() {
		input = parent{Name: "Hello World!", Child: child{Name: "Bob!"}}
		statusCode = http.StatusOK
		rec = httptest.NewRecorder()

		Convey("When the encoder is invoked", func() {
			WriteJSON(rec, input, http.StatusOK)

			Convey("Then the input value is written to the response body.", func() {
				var actual parent
				json.Unmarshal(rec.Body.Bytes(), &actual)
				So(actual, ShouldResemble, input)

				Convey("The response http status code is correct.", func() {
					So(rec.Code, ShouldEqual, statusCode)

					Convey("The response content type header is 'application/json'", func() {
						So(rec.Header().Get(contentTypeHeader), ShouldEqual, contentTypeJSON)

						Convey("And the JSON encoder is called with the expected args the correct number of times.", func() {
							So(len(mock.invocationArgs), ShouldEqual, 1)
							So(mock.invocationArgs[0], ShouldResemble, input)
						})
					})
				})
			})

		})
	})



	/*		Convey("When the encoder is invoked", t, func() {
				WriteJSON(rec, input, http.StatusOK)

				Convey("The response body is as expected", t, func() {
					var actual parent
					json.Unmarshal(rec.Body.Bytes(), &actual)
					So(actual, ShouldResemble, input)
				})
			})
		})*/
}

/*	Convey("Should Write value to response as JSON, set the expected http response status code & set the content "+
		"type header with the expected value", t, func() {
		rec := httptest.NewRecorder()
		mock := mockEncoderInit()
		jsonResponseEncoder = mock

		input := parent{Name: "Hello World!", Child: child{Name: "Bob!"}}

		WriteJSON(rec, input, http.StatusOK)

		var actual parent
		json.Unmarshal(rec.Body.Bytes(), &actual)

		So(actual, ShouldResemble, input)
		So(rec.Code, ShouldEqual, http.StatusOK)
		So(rec.Header().Get(contentTypeHeader), ShouldEqual, contentTypeJSON)
		So(len(mock.invocationArgs), ShouldEqual, 1)
		So(mock.invocationArgs[0], ShouldResemble, input)
	})

	Convey("Should return internal server error http status if the writer cannot successfully "+
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
	})*/
