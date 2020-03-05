package rediscluster

import (
	"time"
)

// cache is an implementation of httpcache.Cache that caches responses in a
// redis cluster.
type cache struct {
	cluster clusterClient
	ttl     time.Duration
}

// cacheKey modifies an httpcache key for use in redis. Specifically, it
// prefixes keys to avoid collision with other data stored in redis.
func cacheKey(key string) string {
	return "rediscache:" + key
}

// Get returns the response corresponding to key if present.
func (c *cache) Get(key string) (resp []byte, ok bool) {
	item, err := c.cluster.Get(cacheKey(key)).Result()
	if err != nil {
		return nil, false
	}
	return []byte(item), true
}

// Set saves a response to the cache as key.
func (c *cache) Set(key string, value []byte) {
	c.cluster.Set(cacheKey(key), string(value), c.ttl)
}

// Delete removes the response with key from the cache.
func (c *cache) Delete(key string) {
	c.cluster.Del(cacheKey(key))
}
