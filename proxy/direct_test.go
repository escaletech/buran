package proxy

import (
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDirect(t *testing.T) {
	Convey("Direct", t, func() {
		Convey("directs requests to backend", func() {
			req := httptest.NewRequest("GET", "http://mysite.com/some/file.js", nil)

			// Act
			Direct("http://new-backend.com").Director(req)

			// Assert
			So(req.URL.String(), ShouldEqual, "http://new-backend.com/some/file.js")
			So(req.Host, ShouldEqual, "new-backend.com")
		})

		Convey("removes cdn portion of Prismic hosts", func() {
			req := httptest.NewRequest("GET", "http://mysite.com/some/file.js", nil)

			// Act
			Direct("http://my-repo.prismic.io").Director(req)

			// Assert
			So(req.URL.String(), ShouldEqual, "http://my-repo.prismic.io/some/file.js")
			So(req.Host, ShouldEqual, "my-repo.prismic.io")
		})

		Convey("uses backend protocol", func() {
			req := httptest.NewRequest("GET", "http://mysite.com/some/file.js", nil)

			// Act
			Direct("https://new-backend.com").Director(req)

			// Assert
			So(req.URL.String(), ShouldEqual, "https://new-backend.com/some/file.js")
		})
	})
}
