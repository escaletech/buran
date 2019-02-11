package proxy

import (
	"net/http"

	"github.com/escaleseo/prismic-proxy-cache/env"
	"github.com/gregjones/httpcache"
)

func NewManager(config env.Config, cache httpcache.Cache) (*ProxyManager, error) {
	return &ProxyManager{
		Root:      newRootHandler(config.BackendURL, cache),
		Documents: newDocumentsHandler(config.BackendURL, cache),
	}, nil
}

type ProxyManager struct {
	Root      http.Handler
	Documents http.Handler
}

func newRootHandler(backendURL string, cache httpcache.Cache) http.Handler {
	cachingTransport := httpcache.NewTransport(cache)
	cachingTransport.Transport = newRootTransport(http.DefaultTransport, backendURL)
	httpClient := &http.Client{
		Transport: cachingTransport,
	}

	return newProxy(newRequestBuilder(backendURL, true), httpClient.Do, forwardResponse)
}

func newDocumentsHandler(backendURL string, cache httpcache.Cache) http.Handler {
	cachingTransport := httpcache.NewTransport(cache)
	httpClient := &http.Client{Transport: cachingTransport}

	return newProxy(newRequestBuilder(backendURL, false), httpClient.Do, forwardResponse)
}
