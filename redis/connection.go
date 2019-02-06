package redis

import (
	"github.com/gomodule/redigo/redis"
)

var pool *redis.Pool

func Connect(redisURL string) {
	pool = &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(redisURL)
		},
	}
}

func Get() redis.Conn {
	return pool.Get()
}
