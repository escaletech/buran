package redis

import (
	"fmt"
	"testing"

	"github.com/rdumont/loki"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRedisCacheProvider(t *testing.T) {
	Convey("Invalidate", t, func() {
		newProvider := func(keyPrefix string) (*RedisCacheProvider, *fakeRedis) {
			redis := &fakeRedis{}
			return &RedisCacheProvider{
				func() redisCommander { return redis },
				keyPrefix,
				10,
			}, redis
		}

		Convey("invalidates many matching keys", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []interface{}{"a", "b"}

			provider, redis := newProvider(keyPrefix)
			redis.DoCalls.On("KEYS", keyPrefix).Return(keys, nil)

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldBeNil)

			delCall := redis.DoCalls.GetNthCall(1)
			So(delCall, ShouldResemble, loki.Params{"DEL", "a", "b"})
		})

		Convey("does nothing for zero matching keys", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []interface{}{}

			provider, redis := newProvider(keyPrefix)
			redis.DoCalls.On("KEYS", keyPrefix).Return(keys, nil)

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldBeNil)
			So(redis.DoCalls.CallCount(), ShouldEqual, 1)
		})

		Convey("returns error when redis response is of incorrect type", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := "keys should not be string"

			provider, redis := newProvider(keyPrefix)
			redis.DoCalls.On("KEYS", keyPrefix).Return(keys, nil)

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "unexpected type string for keys response")
		})

		Convey("returns error from redis keys call", func() {
			// Arrange
			keyPrefix := "key-prefix"

			provider, redis := newProvider(keyPrefix)
			redis.DoCalls.On("KEYS", keyPrefix).Return(nil, fmt.Errorf("crash"))

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "crash")
		})

		Convey("returns error from redis del call", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []interface{}{"a", "b"}

			provider, redis := newProvider(keyPrefix)
			redis.DoCalls.On("KEYS", keyPrefix).Return(keys, nil)
			redis.DoCalls.On("DEL", "a", "b").Return(nil, fmt.Errorf("crash"))

			// Act
			err := provider.Invalidate()

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
