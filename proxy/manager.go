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
	rootHandler      http.Handler
	documentsHandler http.Handler

	redis redigo.Conn
}

func (h *ProxyManager) ServeRoot(w http.ResponseWriter, r *http.Request) {
	h.rootHandler.ServeHTTP(w, r)
}

func (h *ProxyManager) ServeDocuments(w http.ResponseWriter, r *http.Request) {
	h.documentsHandler.ServeHTTP(w, r)
}

func (h *ProxyManager) init(config env.Config) {
	h.redis = redis.Get()
	h.rootHandler = newRootHandler(config.BackendURL, h.redis)
	h.documentsHandler = newDocumentsHandler(config.BackendURL, h.redis)
}

func (h *ProxyManager) startReconnection(config env.Config) {
	disconnected := false
	for {
		time.Sleep(2 * time.Second)
		if _, err := h.redis.Do("PING"); err != nil {
			disconnected = true
			log.WithError(err).Warn("redis connection error, retrying...")
			h.init(config)
		} else if disconnected {
			log.Info("redis connection restored!")
			disconnected = false
		}
	}
}
