package proxy

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRootTransport(t *testing.T) {
	Convey("RoundTrip", t, func() {
		Convey("transforms returned body", func() {
			t := &rootTransport{
				transport:     roundTripReturningURL(http.StatusOK, "received: "),
				transformBody: prependBody("transformed: "),
			}

			// Act
			res, err := t.RoundTrip(httptest.NewRequest("GET", "http://target.com", nil))

			// Assert
			So(err, ShouldBeNil)
			body, _ := ioutil.ReadAll(res.Body)
			So(string(body), ShouldEqual, "transformed: received: http://target.com")
		})

		Convey("adds cache control header", func() {
			for _, code := range []int{http.StatusOK, http.StatusNotModified} {
				Convey("when status code is "+http.StatusText(code), func() {
					t := &rootTransport{
						transport:     roundTripReturningURL(code, ""),
						transformBody: prependBody(""),
					}

					// Act
					res, err := t.RoundTrip(httptest.NewRequest("GET", "http://target.com", nil))

					// Assert
					So(err, ShouldBeNil)
					So(res.Header.Get("Cache-Control"), ShouldEqual, "max-age=604800")
				})
			}
		})

		Convey("doesn't change cache control header", func() {
			for _, code := range []int{199, http.StatusBadRequest} {
				Convey("when status code is "+http.StatusText(code), func() {
					t := &rootTransport{
						transport:     roundTripReturningURL(code, ""),
						transformBody: prependBody(""),
					}

					// Act
					res, err := t.RoundTrip(httptest.NewRequest("GET", "http://target.com", nil))

					// Assert
					So(err, ShouldBeNil)
					So(res.Header.Get("Cache-Control"), ShouldEqual, "")
				})
			}
		})

		Convey("returns error from inner transport", func() {
			t := &rootTransport{
				transport:     roundTripWithError(fmt.Errorf("crash transport")),
				transformBody: prependBody("transformed: "),
			}

			// Act
			_, err := t.RoundTrip(httptest.NewRequest("GET", "http://target.com", nil))

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "crash transport")
		})

		Convey("returns error from body transformation", func() {
			t := &rootTransport{
				transport:     roundTripReturningURL(http.StatusOK, "received: "),
				transformBody: errorBody(fmt.Errorf("crash body")),
			}

			// Act
			_, err := t.RoundTrip(httptest.NewRequest("GET", "http://target.com", nil))

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "crash body")
		})
	})

	Convey("hostReplacer", t, func() {
		const backendURL = "http://the-backend.com/api/v2"

		originalBody := newBody(fmt.Sprintf(`{
			"some key": "%v",
			"deep": {
				"here": "backend is %v!"
			}
		}`, backendURL, backendURL))

		req := httptest.NewRequest("GET", fmt.Sprintf("http://foobar.com?%v=https://replaced", HostParamKey), nil)

		// Act
		newBody, err := hostReplacer(backendURL)(originalBody, req)

		// Assert
		So(err, ShouldBeNil)
		body, _ := ioutil.ReadAll(newBody)
		So(string(body), ShouldEqual, `{
			"some key": "https://replaced",
			"deep": {
				"here": "backend is https://replaced!"
			}
		}`)
	})
}

type fakeRoundTripper struct {
	roundTrip func(req *http.Request) (*http.Response, error)
}

func (rt *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.roundTrip(req)
}

func roundTripWithError(err error) http.RoundTripper {
	return &fakeRoundTripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			return nil, err
		},
	}
}

func roundTripReturningURL(statusCode int, bodyPrefix string) http.RoundTripper {
	return &fakeRoundTripper{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				Header:     http.Header{},
				StatusCode: statusCode,
				Body:       newBody(bodyPrefix + req.URL.String()),
			}, nil
		},
	}
}

func prependBody(prefix string) bodyTransformation {
	return func(body io.ReadCloser, req *http.Request) (io.ReadCloser, error) {
		oldBody, _ := ioutil.ReadAll(body)
		return newBody(prefix + string(oldBody)), nil
	}
}

func errorBody(err error) bodyTransformation {
	return func(body io.ReadCloser, req *http.Request) (io.ReadCloser, error) {
		return nil, err
	}
}
