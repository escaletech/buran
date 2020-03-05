package cache

import (
	"github.com/gregjones/httpcache"
	"github.com/pkg/errors"

	"github.com/escaleseo/buran/internal/cache/memory"
	"github.com/escaleseo/buran/internal/cache/redis"
	"github.com/escaleseo/buran/internal/cache/rediscluster"
	"github.com/escaleseo/buran/internal/platform/env"
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

	case "redis-cluster":
		return rediscluster.New(config)

	default:
		return nil, errors.New("unknown cache provider " + config.CacheProvider)
	}
}
