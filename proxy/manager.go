package proxy

import (
	"net/http"
	"time"

	redigo "github.com/gomodule/redigo/redis"

	"github.com/escaleseo/prismic-proxy-cache/env"
	"github.com/escaleseo/prismic-proxy-cache/logger"
	"github.com/escaleseo/prismic-proxy-cache/redis"
)

var log = logger.Get()

func NewManager(config env.Config) (*ProxyManager, error) {
	handler := new(ProxyManager)
	handler.init(config)
	go handler.startReconnection(config)
	return handler, nil
}

type ProxyManager struct {
	Root      http.Handler
	Documents http.Handler

	redis redigo.Conn
}

func (h *ProxyManager) init(config env.Config) {
	h.redis = redis.Get()
	h.Root = newRootHandler(config.BackendURL, h.redis)
	h.Documents = newDocumentsHandler(config.BackendURL, h.redis)
}

func (h *ProxyManager) startReconnection(config env.Config) {
	for {
		time.Sleep(2 * time.Second)
		if _, err := h.redis.Do("PING"); err != nil {
			log.WithError(err).Warn("redis connection error, retrying...")
			h.init(config)
		}
	}
}
