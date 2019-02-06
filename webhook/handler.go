package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/escaleseo/prismic-proxy-cache/env"
	"github.com/escaleseo/prismic-proxy-cache/logger"
	"github.com/escaleseo/prismic-proxy-cache/proxy"
	"github.com/escaleseo/prismic-proxy-cache/redis"
)

var log = logger.Get()

func New(config env.Config) http.Handler {
	return &webhookHandler{
		cacheKey:   fmt.Sprintf("rediscache:%v/api/v2?%v=*", config.BackendURL, proxy.HostParamKey),
		invalidate: invalidateAPI,
		getRedis:   func() redisCommander { return redis.Get() },
	}
}

type webhookHandler struct {
	cacheKey   string
	invalidate invalidate
	getRedis   func() redisCommander
}

func (h *webhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type string `json:"type,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if body.Type != "api-update" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unknown webhook type " + body.Type))
		return
	}

	if err := h.invalidate(h.getRedis(), h.cacheKey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func invalidateAPI(redis redisCommander, keyPrefix string) error {
	res, err := redis.Do("KEYS", keyPrefix)
	if err != nil {
		return err
	}

	keys, ok := res.([]interface{})
	if !ok {
		return fmt.Errorf("unexpected type %v for keys response", reflect.TypeOf(res))
	}

	if len(keys) > 0 {
		_, err = redis.Do("DEL", keys...)
	}

	return err
}

type redisCommander interface {
	Do(command string, args ...interface{}) (reply interface{}, err error)
}

type invalidate func(redis redisCommander, keyPrefix string) error
