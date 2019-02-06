package main

import (
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/escaleseo/prismic-proxy-cache/env"
	"github.com/escaleseo/prismic-proxy-cache/logger"
	"github.com/escaleseo/prismic-proxy-cache/proxy"
	"github.com/escaleseo/prismic-proxy-cache/redis"
	"github.com/escaleseo/prismic-proxy-cache/webhook"
)

func main() {
	config := env.GetConfig()
	redis.Connect(config.RedisURL)

	log := logger.Get()

	proxies, err := proxy.NewManager(config)
	if err != nil {
		log.WithError(err).Fatal("failed to create proxy handler")
	}

	webhookHandler := webhook.New(config)

	router := mux.NewRouter()
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	router.Use(logger.NewMiddleware())
	router.Handle("/_webhook", webhookHandler)
	router.Path("/api/v2").Handler(proxies.Root)
	router.PathPrefix("/api/v2/documents").Handler(proxies.Documents)
	router.NewRoute().Handler(proxyTo("escale-health.prismic.io"))

	log.Info("listening on port ", config.Port)
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		log.WithError(err).Error("server quit")
	}
}

func proxyTo(host string) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.Header.Set("X-Forwarded-Host", req.Host)

		req.Host = host
		req.URL.Host = req.Host
		req.URL.Scheme = "http"
	}
	return &httputil.ReverseProxy{Director: director}
}
