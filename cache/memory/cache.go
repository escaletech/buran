package memory

import (
	"strings"
	"sync"

	"github.com/gregjones/httpcache"
)

type cache struct {
	httpcache.Cache
	mu       sync.Mutex
	rootKeys []string
}

func (c *cache) Set(key string, responseBytes []byte) {
	c.Cache.Set(key, responseBytes)
	if strings.Contains(key, "/api/v2?") {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.rootKeys = append(c.rootKeys, key)
	}
}

func (c *cache) DeleteAllRootKeys() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range c.rootKeys {
		c.Cache.Delete(k)
	}
	c.rootKeys = nil
}
