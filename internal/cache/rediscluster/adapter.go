package rediscluster

import (
	"sync"

	"github.com/go-redis/redis"
)

type clusterAdapter struct {
	*redis.ClusterClient
}

func (a *clusterAdapter) Keys(pattern string) *redis.StringSliceCmd {
	keyMap := map[string]struct{}{}
	var lock sync.Mutex
	err := a.ForEachNode(func(client *redis.Client) error {
		keys, err := client.Keys(pattern).Result()
		if err != nil {
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		for _, k := range keys {
			keyMap[k] = struct{}{}
		}

		return nil
	})

	keys := make([]string, len(keyMap))
	i := 0
	for k := range keyMap {
		keys[i] = k
		i++
	}

	return redis.NewStringSliceResult(keys, err)
}
