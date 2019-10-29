package redis

import (
	"strconv"

	"github.com/escaleseo/buran/env"
	"github.com/gomodule/redigo/redis"
)

// cache is an implementation of httpcache.Cache that caches responses in a
// redis server.
type cache struct {
	conn connectionGetter
}

// cacheKey modifies an httpcache key for use in redis. Specifically, it
// prefixes keys to avoid collision with other data stored in redis.
func cacheKey(key string) string {
	return "rediscache:" + key
}

// Get returns the response corresponding to key if present.
func (c *cache) Get(key string) (resp []byte, ok bool) {
	item, err := redis.Bytes(c.conn().Do("GET", cacheKey(key)))
	if err != nil {
		return nil, false
	}
	return item, true
}

// Set saves a response to the cache as key.
func (c *cache) Set(key string, value []byte) {
	ttl, _ := strconv.Atoi(env.GetConfig().TTL)
	c.conn().Do("SET", cacheKey(key), value, "EX", ttl)
}

// Delete removes the response with key from the cache.
func (c *cache) Delete(key string) {
	c.conn().Do("DEL", cacheKey(key))
}
