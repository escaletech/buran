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
				10 * time.Second,
			}, cluster
		}

		Convey("invalidates many matching keys", func() {
			// Arrange
			keyPrefix := "key-prefix"

			provider, cluster := newProvider(keyPrefix)
			cluster.DelKeysCalls.On(keyPrefix).Return(nil)

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldBeNil)

			delCall := cluster.DelKeysCalls.GetNthCall(0)
			So(delCall, ShouldResemble, loki.Params{keyPrefix})
		})

		Convey("returns error from delete keys call", func() {
			// Arrange
			keyPrefix := "key-prefix"

			provider, cluster := newProvider(keyPrefix)
			cluster.DelKeysCalls.On(keyPrefix).Return(fmt.Errorf("crash"))

			// Act
			err := provider.Invalidate()

			// Assert
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "crash")
		})
	})
}

type fakeRedis struct {
	KeysCalls    loki.Method
	DelCalls     loki.Method
	DelKeysCalls loki.Method
	GetCalls     loki.Method
	SetCalls     loki.Method
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

func (r *fakeRedis) DelKeys(pattern string) error {
	err, _ := r.DelKeysCalls.Receive(pattern).GetOr(0, nil).(error)
	return err
}
