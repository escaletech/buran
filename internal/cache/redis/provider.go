package redis

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/escaleseo/buran/internal/proxy"
	"github.com/pkg/errors"

	"github.com/gomodule/redigo/redis"
	"github.com/gregjones/httpcache"

	"github.com/escaleseo/buran/internal/platform/env"
)

type redisCommander interface {
	Do(command string, args ...interface{}) (reply interface{}, err error)
}

type connectionGetter func() redisCommander

func New(config env.Config) (*RedisCacheProvider, error) {
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(config.RedisURL)
		},
	}

	ttl, err := strconv.Atoi(config.TTL)
	if err != nil {
		return nil, errors.Wrap(err, "invalid ttl")
	}

	return &RedisCacheProvider{
		conn:       func() redisCommander { return pool.Get() },
		keyPattern: keyPattern(config.BackendURL),
		ttl:        ttl,
	}, nil
}

type RedisCacheProvider struct {
	conn       connectionGetter
	keyPattern string
	ttl        int
}

func (p *RedisCacheProvider) Invalidate() error {
	redis := p.conn()
	res, err := redis.Do("KEYS", p.keyPattern)
	if err != nil {
		return err
	}

	keys, ok := res.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected type %v for keys response", reflect.TypeOf(res))
	}

	if len(keys) > 0 {
		_, err = redis.Do("DEL", keys...)
	}

	return err
}

func (p *RedisCacheProvider) GetCache() httpcache.Cache {
	return &cache{p.conn, p.ttl}
}

func keyPattern(backendURL string) string {
	return fmt.Sprintf("rediscache:%v/api/v2?%v=*", backendURL, proxy.HostParamKey)
}
