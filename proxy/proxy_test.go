package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProxy(t *testing.T) {
	Convey("newProxy", t, func() {
		Convey("forwards response in the end", func() {
			proxy := newProxy(
				buildRequestReplacingHost("target.com"),
				doRequestReturningURL(http.StatusOK, "called: "),
				forwardStatusCode)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://original.com", nil)

			// Act
			proxy(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "called: http://target.com")
		})

		Convey("returns error if request fails to build", func() {
			proxy := newProxy(
				buildRequestWithError(fmt.Errorf("crash request builder")),
				doRequestReturningURL(http.StatusOK, "called: "),
				forwardStatusCode)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://original.com", nil)

			// Act
			proxy(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body.String(), ShouldEqual, "crash request builder")
		})

		Convey("returns error if request fails", func() {
			proxy := newProxy(
				buildRequestReplacingHost("target.com"),
				doRequestWithError(fmt.Errorf("crash request")),
				forwardStatusCode)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://original.com", nil)

			// Act
			proxy(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body.String(), ShouldEqual, "crash request")
		})
	})

	Convey("forwardResponse", t, func() {
		Convey("copies header, status code and body", func() {
			w := httptest.NewRecorder()
			header := http.Header{"X-Custom": []string{"ok"}}
			res := &http.Response{
				StatusCode: 666,
				Header:     header,
				Body:       newBody("Hello, World!"),
			}

			// Act
			err := forwardResponse(w, res)

			// Assert
			So(err, ShouldBeNil)
			So(w.Code, ShouldEqual, res.StatusCode)
			So(w.HeaderMap, ShouldResemble, header)
			So(w.Body.String(), ShouldEqual, "Hello, World!")
		})
	})

	Convey("newRequestBuilder", t, func() {
		always := func(isRootAPI bool) {
			buildRequest := newRequestBuilder("http://target.com/some/prefix", isRootAPI)

			Convey("composes basic request", func() {
				r := createSourceRequest("http://mysite.com/forward-this", nil)

				req, err := buildRequest(r)

				So(err, ShouldBeNil)
				So(req.URL.String(), ShouldStartWith, "http://target.com/some/prefix/forward-this")
			})

			Convey("trims trailing slash from URL", func() {
				r := createSourceRequest("http://mysite.com/forward-this/", nil)

				req, err := buildRequest(r)

				So(err, ShouldBeNil)
				So(req.URL.String(), ShouldStartWith, "http://target.com/some/prefix/forward-this")
			})

			Convey("fails to create requests with bad base URL", func() {
				buildRequest := newRequestBuilder("%=not-an-url", isRootAPI)
				r := createSourceRequest("http://mysite.com/forward-this", nil)

				_, err := buildRequest(r)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "%=not-an-url")
			})
		}

		Convey("for root API", func() {
			buildRequest := newRequestBuilder("http://target.com/some/prefix", true)

			Convey("appends proxy host parameter", func() {
				r := createSourceRequest("http://mysite.com/forward-this", nil)

				req, err := buildRequest(r)

				So(err, ShouldBeNil)
				So(req.URL.String(), ShouldEqual, "http://target.com/some/prefix/forward-this?proxy-host=http%3A%2F%2Fmysite.com")
			})

			Convey("appends proxy host parameter with HTTPS", func() {
				r := createSourceRequest("http://mysite.com/forward-this", nil)
				r.Header.Set("X-Forwarded-Proto", "https")

				req, err := buildRequest(r)

				So(err, ShouldBeNil)
				So(req.URL.String(), ShouldEqual, "http://target.com/some/prefix/forward-this?proxy-host=https%3A%2F%2Fmysite.com")
			})

			Convey("forwards headers that are not black-listed", func() {
				r := createSourceRequest("http://mysite.com/forward-this", map[string]string{
					"X-Custom":      "some-value",
					"Cache-Control": "no-cache",
				})

				req, err := buildRequest(r)

				So(err, ShouldBeNil)
				So(req.Header, ShouldResemble, http.Header{
					"X-Custom": []string{"some-value"},
				})
			})

			always(true)
		})

		Convey("for documents API", func() {
			buildRequest := newRequestBuilder("http://target.com/some/prefix", false)

			Convey("forwards all query string parameters", func() {
				r := createSourceRequest("http://mysite.com/forward-this?paramOne=foo&second=bar", nil)

				req, err := buildRequest(r)

				So(err, ShouldBeNil)
				So(req.URL.String(), ShouldEqual, "http://target.com/some/prefix/forward-this?paramOne=foo&second=bar")
			})

			Convey("forwards all headers", func() {
				r := createSourceRequest("http://mysite.com/forward-this", map[string]string{
					"X-Custom":      "some-value",
					"Cache-Control": "no-cache",
				})

				req, err := buildRequest(r)

				So(err, ShouldBeNil)
				So(req.Header, ShouldResemble, http.Header{
					"X-Custom":      []string{"some-value"},
					"Cache-Control": []string{"no-cache"},
				})
			})

			always(false)
		})
	})
}

func createSourceRequest(url string, headers map[string]string) *http.Request {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic("invalid request: " + err.Error())
	}

	for k, v := range headers {
		r.Header.Set(k, v)
	}

	return r
}

func newBody(text string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(text)))
}

func newJSONBody(body interface{}) io.ReadCloser {
	marshalled, _ := json.Marshal(body)
	return newBody(string(marshalled))
}

func buildRequestReplacingHost(newHost string) requestBuilder {
	return func(r *http.Request) (*http.Request, error) {
		return httptest.NewRequest(r.Method, strings.Replace(r.URL.String(), r.URL.Host, newHost, -1), nil), nil
	}
}

func buildRequestWithError(err error) requestBuilder {
	return func(r *http.Request) (*http.Request, error) {
		return nil, err
	}
}

func doRequestReturningURL(statusCode int, bodyPrefix string) httpRequester {
	return func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: statusCode,
			Header:     http.Header{},
			Body:       newBody(bodyPrefix + r.URL.String()),
		}, nil
	}
}

func doRequestWithError(err error) httpRequester {
	return func(r *http.Request) (*http.Response, error) {
		return nil, err
	}
}

func forwardStatusCode(w http.ResponseWriter, res *http.Response) error {
	w.WriteHeader(res.StatusCode)
	body, _ := ioutil.ReadAll(res.Body)
	w.Write(body)
	return nil
}
