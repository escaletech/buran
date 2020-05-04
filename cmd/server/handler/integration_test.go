// +build integration

package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/escaletech/buran/cmd/server/handler"
	"github.com/escaletech/buran/internal/platform/env"
	"github.com/escaletech/buran/internal/platform/logger"
	"github.com/kelseyhightower/envconfig"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gentleman.v2"
)

func TestHandler(t *testing.T) {
	var config env.Config
	envconfig.MustProcess("", &config)

	getEnv := func(t *testing.T) (*TestEnvironment, func()) {
		te := newEnvironment(t, config)
		return te, func() {
			te.Server.Close()
			te.BackendServer.Close()
		}
	}

	Convey(config.CacheProvider+" cache provider", t, func() {
		Convey("root API call", func() {
			Convey("returns fresh result", func() {
				te, tearDown := getEnv(t)
				defer tearDown()

				be, reqs := handleStatic(200, "ok")
				te.BackendHandler = be

				// Act
				res, err := te.Client.Get().AddPath("/api/v2").Do()

				// Assert
				So(err, ShouldBeNil)
				So(res.StatusCode, ShouldEqual, 200)
				So(res.String(), ShouldEqual, "ok")
				So(len(*reqs), ShouldEqual, 1)
			})

			Convey("returns cached result without repeating request", func() {
				te, tearDown := getEnv(t)
				defer tearDown()

				be, reqs := handleStatic(200, "ok")
				te.BackendHandler = be

				te.Client.Get().AddPath("/api/v2").Do()

				// Act
				res, err := te.Client.Get().AddPath("/api/v2").Do()

				// Assert
				So(err, ShouldBeNil)
				So(res.StatusCode, ShouldEqual, 200)
				So(res.String(), ShouldEqual, "ok")
				So(len(*reqs), ShouldEqual, 1)
			})

			Convey("returns fresh again after receiving webhook call", func() {
				te, tearDown := getEnv(t)
				defer tearDown()

				be, reqs := handleStatic(200, "ok")
				te.BackendHandler = be

				te.Client.Get().AddPath("/api/v2").Do()
				te.CallWebhook()

				// Act
				res, err := te.Client.Get().AddPath("/api/v2").Do()

				// Assert
				So(err, ShouldBeNil)
				So(res.StatusCode, ShouldEqual, 200)
				So(res.String(), ShouldEqual, "ok")
				So(len(*reqs), ShouldEqual, 2)
			})

			Convey("makes a new call for each host header variation", func() {
				te, tearDown := getEnv(t)
				defer tearDown()

				be, reqs := handleStatic(200, "ok")
				te.BackendHandler = be

				// Act
				const reps = 10
				for i := 0; i < reps; i++ {
					te.Client.Get().AddPath("/api/v2").
						SetHeader("X-Forwarded-Host", fmt.Sprintf("host-%v", i)).
						Do()
				}

				So(len(*reqs), ShouldEqual, reps)
			})

			Convey("webhook invalidates every host variation", func() {
				te, tearDown := getEnv(t)
				defer tearDown()

				be, reqs := handleStatic(200, "ok")
				te.BackendHandler = be

				const reps = 10
				for i := 0; i < reps; i++ {
					te.Client.Get().AddPath("/api/v2").
						SetHeader("X-Forwarded-Host", fmt.Sprintf("host-%v", i)).
						Do()
				}

				te.CallWebhook()

				// Act
				for i := 0; i < reps; i++ {
					te.Client.Get().AddPath("/api/v2").
						SetHeader("X-Forwarded-Host", fmt.Sprintf("host-%v", i)).
						Do()
				}

				So(len(*reqs), ShouldEqual, reps*2)
			})
		})
	})
}

func handleStatic(code int, body string) (http.HandlerFunc, *[]*http.Request) {
	reqs := &[]*http.Request{}
	lock := sync.Mutex{}
	h := func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()

		*reqs = append(*reqs, r)

		w.WriteHeader(code)
		w.Write([]byte(body))
	}
	return h, reqs
}

func newEnvironment(t *testing.T, config env.Config) *TestEnvironment {
	log := logger.Get()
	log.Out = ioutil.Discard

	te := &TestEnvironment{}

	backend := func(w http.ResponseWriter, r *http.Request) {
		if te.BackendHandler == nil {
			t.Error("no backend handler configured")
			t.FailNow()
		}
		te.BackendHandler(w, r)
	}
	te.BackendServer = httptest.NewServer(http.HandlerFunc(backend))
	config.BackendURL = te.BackendServer.URL

	h, err := handler.New(config, log)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	te.Server = httptest.NewServer(h)
	te.Client = gentleman.New().BaseURL(te.Server.URL)

	return te
}

type TestEnvironment struct {
	Server         *httptest.Server
	Client         *gentleman.Client
	BackendHandler http.HandlerFunc
	BackendServer  *httptest.Server
}

func (te *TestEnvironment) CallWebhook() {
	res, err := te.Client.Post().AddPath("/_webhook").
		JSON(map[string]interface{}{"type": "api-update"}).
		Do()

	So(err, ShouldBeNil)
	So(res.StatusCode, ShouldEqual, 200)
	if res.StatusCode != 200 {
		Println("webhook response: ", res.String())
	}
}
