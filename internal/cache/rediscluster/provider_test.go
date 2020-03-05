package rediscluster

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/rdumont/loki"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRedisClusterCacheProvider(t *testing.T) {
	Convey("Invalidate", t, func() {
		newProvider := func(keyPrefix string) (*RedisClusterCacheProvider, *fakeRedis) {
			cluster := &fakeRedis{}
			return &RedisClusterCacheProvider{
				cluster,
				keyPrefix,
			}, cluster
		}

		Convey("invalidates many matching keys", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []string{"a", "b"}

			provider, cluster := newProvider(keyPrefix)
			cluster.KeysCalls.On(keyPrefix).Return(redis.NewStringSliceResult(keys, nil))

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldBeNil)

			delCall := cluster.DelCalls.GetNthCall(0)
			So(delCall, ShouldResemble, loki.Params{"a", "b"})
		})

		Convey("does nothing for zero matching keys", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []string{}

			provider, cluster := newProvider(keyPrefix)
			cluster.KeysCalls.On(keyPrefix).Return(redis.NewStringSliceResult(keys, nil))

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldBeNil)
			So(cluster.DelCalls.CallCount(), ShouldEqual, 0)
		})

		Convey("returns error from redis keys call", func() {
			// Arrange
			keyPrefix := "key-prefix"

			provider, cluster := newProvider(keyPrefix)
			cluster.KeysCalls.On(keyPrefix).
				Return(redis.NewStringSliceResult(nil, fmt.Errorf("crash")))

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "crash")
		})

		Convey("returns error from redis del call", func() {
			// Arrange
			keyPrefix := "key-prefix"
			keys := []string{"a", "b"}

			provider, cluster := newProvider(keyPrefix)
			cluster.KeysCalls.On(keyPrefix).Return(redis.NewStringSliceResult(keys, nil))
			cluster.DelCalls.On("a", "b").Return(redis.NewIntResult(0, fmt.Errorf("crash")))

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "crash")
		})
	})
}

type fakeRedis struct {
	KeysCalls loki.Method
	DelCalls  loki.Method
	GetCalls  loki.Method
	SetCalls  loki.Method
}

func (r *fakeRedis) Keys(pattern string) *redis.StringSliceCmd {
	return r.KeysCalls.Receive(pattern).GetOr(0, new(redis.StringSliceCmd)).(*redis.StringSliceCmd)
}

func (r *fakeRedis) Del(keys ...string) *redis.IntCmd {
	args := make([]interface{}, len(keys))
	for i, k := range keys {
		args[i] = k
	}
	return r.DelCalls.Receive(args...).GetOr(0, new(redis.IntCmd)).(*redis.IntCmd)
}

func (r *fakeRedis) Get(key string) *redis.StringCmd {
	return r.GetCalls.Receive(key).GetOr(0, new(redis.StringCmd)).(*redis.StringCmd)
}

func (r *fakeRedis) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.SetCalls.Receive(key, value).GetOr(0, new(redis.StatusCmd)).(*redis.StatusCmd)
}
