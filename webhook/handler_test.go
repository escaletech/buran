package webhook

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rdumont/loki"
	. "github.com/smartystreets/goconvey/convey"
)

func TestWebhookHandler(t *testing.T) {
	Convey("ServeHTTP", t, func() {
		noRedis := func() redisCommander { return nil }

		newRequest := func(body string) *http.Request {
			return httptest.NewRequest("GET", "/", bytes.NewReader([]byte(body)))
		}

		newInvalidator := func(prefix *string, err error) invalidate {
			return func(redis redisCommander, keyPrefix string) error {
				*prefix = keyPrefix
				return err
			}
		}

		Convey("invalidates cache and returns ok", func() {
			// Arrange
			var invalidatedPrefix string
			invalidate := newInvalidator(&invalidatedPrefix, nil)

			w := httptest.NewRecorder()
			r := newRequest(`{"type": "api-update"}`)

			// Act
			(&webhookHandler{"key-prefix", invalidate, noRedis}).ServeHTTP(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "ok")

			So(invalidatedPrefix, ShouldEqual, "key-prefix")
		})

		Convey("returns bad request if body cannot be read", func() {
			// Arrange
			var invalidatedPrefix string
			invalidate := newInvalidator(&invalidatedPrefix, nil)

			w := httptest.NewRecorder()
			r := newRequest(`{invalid JSON`)

			// Act
			(&webhookHandler{"key-prefix", invalidate, noRedis}).ServeHTTP(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusBadRequest)
			So(w.Body.String(), ShouldStartWith, "invalid character")
			So(invalidatedPrefix, ShouldBeEmpty)
		})

		Convey("skips invalidation and returns ok if hook type is unknown", func() {
			// Arrange
			var invalidatedPrefix string
			invalidate := newInvalidator(&invalidatedPrefix, nil)

			w := httptest.NewRecorder()
			r := newRequest(`{"type": "UNKNOWN"}`)

			// Act
			(&webhookHandler{"key-prefix", invalidate, noRedis}).ServeHTTP(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldEqual, "unknown webhook type UNKNOWN")

			So(invalidatedPrefix, ShouldBeEmpty)
		})

		Convey("returns internal server error if invalidation fails", func() {
			// Arrange
			var invalidatedPrefix string
			invalidate := newInvalidator(&invalidatedPrefix, fmt.Errorf("crash"))

			w := httptest.NewRecorder()
			r := newRequest(`{"type": "api-update"}`)

			// Act
			(&webhookHandler{"key-prefix", invalidate, noRedis}).ServeHTTP(w, r)

			// Assert
			So(w.Code, ShouldEqual, http.StatusInternalServerError)
			So(w.Body.String(), ShouldEqual, "crash\n")
		})
	})

	Convey("invalidateAPI", t, func() {
		Convey("invalidates many matching keys", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []interface{}{"a", "b"}

			redis := new(fakeRedis)
			redis.DoCalls.On("KEYS", keyPrefix).Return(keys, nil)

			// Act
			err := invalidateAPI(redis, keyPrefix)

			// Assert
			So(err, ShouldBeNil)

			delCall := redis.DoCalls.GetNthCall(1)
			So(delCall, ShouldResemble, loki.Params{"DEL", "a", "b"})
		})

		Convey("does nothing for zero matching keys", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []interface{}{}

			redis := new(fakeRedis)
			redis.DoCalls.On("KEYS", keyPrefix).Return(keys, nil)

			// Act
			err := invalidateAPI(redis, keyPrefix)

			// Assert
			So(err, ShouldBeNil)
			So(redis.DoCalls.CallCount(), ShouldEqual, 1)
		})

		Convey("returns error when redis response is of incorrect type", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := "keys should not be string"

			redis := new(fakeRedis)
			redis.DoCalls.On("KEYS", keyPrefix).Return(keys, nil)

			// Act
			err := invalidateAPI(redis, keyPrefix)

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unexpected type string for keys response")
		})

		Convey("returns error from redis keys call", func() {
			// Arrange
			keyPrefix := "key-prefix"

			redis := new(fakeRedis)
			redis.DoCalls.On("KEYS", keyPrefix).Return(nil, fmt.Errorf("crash"))

			// Act
			err := invalidateAPI(redis, keyPrefix)

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "crash")
		})

		Convey("returns error from redis del call", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []interface{}{"a", "b"}

			redis := new(fakeRedis)
			redis.DoCalls.On("KEYS", keyPrefix).Return(keys, nil)
			redis.DoCalls.On("DEL", "a", "b").Return(nil, fmt.Errorf("crash"))

			// Act
			err := invalidateAPI(redis, keyPrefix)

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "crash")
		})
	})
}

type fakeRedis struct {
	DoCalls loki.Method
}

func (r *fakeRedis) Do(command string, args ...interface{}) (interface{}, error) {
	res := r.DoCalls.Receive(append([]interface{}{command}, args...)...)
	reply := res.GetOr(0, nil)
	err, _ := res.GetOr(1, nil).(error)

	return reply, err
}
