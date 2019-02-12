package cache

import (
	"github.com/escaleseo/buran/cache/memory"
	"github.com/escaleseo/buran/cache/redis"
	"github.com/escaleseo/buran/env"
	"github.com/gregjones/httpcache"
	"github.com/pkg/errors"
)

type Provider interface {
	Invalidate() error
	GetCache() httpcache.Cache
}

func NewProvider(config env.Config) (Provider, error) {
	switch config.CacheProvider {
	case "redis":
		return redis.New(config)

	case "memory":
		return memory.New()

	default:
		return nil, errors.New("unknown cache provider " + config.CacheProvider)
	}
}
