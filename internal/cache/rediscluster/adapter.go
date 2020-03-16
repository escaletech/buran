package rediscluster

import (
	"sync"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

type clusterAdapter struct {
	*redis.ClusterClient
}

func (a *clusterAdapter) DelKeys(pattern string) error {
	return a.ForEachMaster(func(client *redis.Client) error {
		keys, err := client.Keys(pattern).Result()
		if err != nil {
			return err
		}

		if len(keys) == 0 {
			return nil
		}

		wg := sync.WaitGroup{}
		for _, k := range keys {
			wg.Add(1)
			go func(k string) {
				defer wg.Done()
				err = client.Del(k).Err()
			}(k)
		}

		wg.Wait()
		if err != nil {
			return errors.Wrap(err, "failed to delete keys")
		}

		return nil
	})
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
