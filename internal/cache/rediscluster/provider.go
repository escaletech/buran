package rediscluster

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/gregjones/httpcache"
	"github.com/pkg/errors"

	"github.com/escaleseo/buran/internal/platform/env"
	"github.com/escaleseo/buran/internal/proxy"
)

type clusterClient interface {
	Keys(pattern string) *redis.StringSliceCmd
	Del(keys ...string) *redis.IntCmd
	Get(key string) *redis.StringCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

func New(config env.Config) (*RedisClusterCacheProvider, error) {
	opts, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, err
	}

	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{opts.Addr},
	})

	ttl, err := strconv.Atoi(env.GetConfig().TTL)
	if err != nil {
		return nil, errors.Wrap(err, "invalid TTL")
	}

	return &RedisClusterCacheProvider{
		cluster:    &clusterAdapter{clusterClient},
		keyPattern: keyPattern(config.BackendURL),
		ttl:        time.Duration(ttl) * time.Second,
	}, nil
}

type RedisClusterCacheProvider struct {
	cluster    clusterClient
	keyPattern string
	ttl        time.Duration
}

func (p *RedisClusterCacheProvider) Invalidate() error {
	keys, err := p.cluster.Keys(p.keyPattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return p.cluster.Del(keys...).Err()
	}

	return err
}

func (p *RedisClusterCacheProvider) GetCache() httpcache.Cache {
	return &cache{p.cluster, p.ttl}
}

func keyPattern(backendURL string) string {
	return fmt.Sprintf("rediscache:%v/api/v2?%v=*", backendURL, proxy.HostParamKey)
}
