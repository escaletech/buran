package webhook

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWebhookHandler(t *testing.T) {
	Convey("ServeHTTP", t, func() {
		newRequest := func(body string) *http.Request {
			return httptest.NewRequest("GET", "/", bytes.NewReader([]byte(body)))
		}

		Convey("invalidates cache and returns ok", func() {
			// Arrange
			invalidated := false
			invalidate := func() error {
				invalidated = true
				return nil
			}

			w := httptest.NewRecorder()
			r := newRequest(`{"type": "api-update"}`)

			// Act
			(&webhookHandler{invalidate}).ServeHTTP(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "ok")

			So(invalidated, ShouldBeTrue)
		})

		Convey("returns bad request if body cannot be read", func() {
			// Arrange
			invalidated := false
			invalidate := func() error {
				invalidated = true
				return nil
			}

			w := httptest.NewRecorder()
			r := newRequest(`{invalid JSON`)

			// Act
			(&webhookHandler{invalidate}).ServeHTTP(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldStartWith, "invalid character")
			So(invalidated, ShouldBeFalse)
		})

		Convey("skips invalidation and returns ok if hook type is unknown", func() {
			// Arrange
			invalidated := false
			invalidate := func() error {
				invalidated = true
				return nil
			}

			w := httptest.NewRecorder()
			r := newRequest(`{"type": "UNKNOWN"}`)

			// Act
			(&webhookHandler{invalidate}).ServeHTTP(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "unknown webhook type UNKNOWN")

			So(invalidated, ShouldBeFalse)
		})

		Convey("returns internal server error if invalidation fails", func() {
			// Arrange
			invalidate := func() error {
				return fmt.Errorf("crash")
			}

			w := httptest.NewRecorder()
			r := newRequest(`{"type": "api-update"}`)

			// Act
			(&webhookHandler{invalidate}).ServeHTTP(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body.String(), ShouldEqual, "crash\n")
		})
	})
}
