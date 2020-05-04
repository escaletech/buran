package handler

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/escaletech/buran/internal/cache"
	"github.com/escaletech/buran/internal/platform/env"
	"github.com/escaletech/buran/internal/platform/logger"
	"github.com/escaletech/buran/internal/proxy"
	"github.com/escaletech/buran/internal/webhook"
)

func New(config env.Config, log *logrus.Logger) (http.Handler, error) {
	cacheProvider, err := cache.NewProvider(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get cache provider")
	}

	proxies, err := proxy.NewManager(config, cacheProvider.GetCache())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create proxy handler")
	}

	webhookHandler := webhook.New(cacheProvider)

	router := mux.NewRouter()
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	router.Use(logger.NewMiddleware())
	router.Handle("/_webhook", webhookHandler)
	router.Path("/api/v2").Handler(proxies.Root)
	router.PathPrefix("/api/v2/documents").Handler(proxies.Documents)
	router.NewRoute().Handler(proxy.Direct(config.BackendURL))

	return router, nil
}
