package memory

import (
	"fmt"

	"github.com/gregjones/httpcache"

	"github.com/escaleseo/buran/internal/proxy"
)

func New() (*MemoryCacheProvider, error) {
	return &MemoryCacheProvider{
		cache: &cache{Cache: httpcache.NewMemoryCache()},
	}, nil
}

type MemoryCacheProvider struct {
	cache *cache
}

func (p *MemoryCacheProvider) Invalidate() error {
	p.cache.DeleteAllRootKeys()
	return nil
}

func (p *MemoryCacheProvider) GetCache() httpcache.Cache {
	return p.cache
}

func keyPattern(backendURL string) string {
	return fmt.Sprintf("rediscache:%v/api/v2?%v=*", backendURL, proxy.HostParamKey)
}
