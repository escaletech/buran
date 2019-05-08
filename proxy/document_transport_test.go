package proxy

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDocumentTransport(t *testing.T) {
	Convey("RoundTrip", t, func() {
		Convey("transforms returned body", func() {
			t := &documentTransport{
				transport:     roundTripReturningBody(http.StatusOK, `"image":{"url":"http://my-fake-image.jpeg"}`),
				transformBody: replaceImagesURLProtocol(),
			}

			// Act
			res, err := t.RoundTrip(httptest.NewRequest("GET", "http://target.com", nil))

			// Assert
			So(err, ShouldBeNil)
			body, _ := ioutil.ReadAll(res.Body)
			So(string(body), ShouldEqual, `"image":{"url":"https://my-fake-image.jpeg"}`)
		})
	})

}

func roundTripReturningBody(statusCode int, body string) http.RoundTripper {
	return &fakeRoundTripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				Header:     http.Header{},
				StatusCode: statusCode,
				Body:       newBody(body),
			}, nil
		},
	}
}
