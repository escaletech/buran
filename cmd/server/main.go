package main

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/escaleseo/buran/internal/cache"
	"github.com/escaleseo/buran/internal/platform/env"
	"github.com/escaleseo/buran/internal/platform/logger"
	"github.com/escaleseo/buran/internal/proxy"
	"github.com/escaleseo/buran/internal/webhook"
)

func main() {
	config := env.GetConfig()

	log := logger.Get()

	cacheProvider, err := cache.NewProvider(config)
	if err != nil {
		log.WithError(err).Fatal("failed do get cache provider")
	}

	proxies, err := proxy.NewManager(config, cacheProvider.GetCache())
	if err != nil {
		log.WithError(err).Fatal("failed to create proxy handler")
	}

	webhookHandler := webhook.New(cacheProvider)

	router := mux.NewRouter()
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	router.Use(logger.NewMiddleware())
	router.Handle("/_webhook", webhookHandler)
	router.Path("/api/v2").Handler(proxies.Root)
	router.PathPrefix("/api/v2/documents").Handler(proxies.Documents)
	router.NewRoute().Handler(proxy.Direct(config.BackendURL))

	log.Info("listening on port ", config.Port)
	if err := http.ListenAndServe(":"+config.Port, router); err != nil {
		log.WithError(err).Error("server quit")
	}
}
