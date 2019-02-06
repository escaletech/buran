package proxy

import (
	"net/http"

	"github.com/gomodule/redigo/redis"
	"github.com/gregjones/httpcache"
	rediscache "github.com/gregjones/httpcache/redis"
)

func newRootHandler(backendURL string, redis redis.Conn) http.Handler {
	cache := rediscache.NewWithClient(redis)
	cachingTransport := httpcache.NewTransport(cache)
	cachingTransport.Transport = newRootTransport(http.DefaultTransport, backendURL)
	httpClient := &http.Client{
		Transport: cachingTransport,
	}

	return newProxy(newRequestBuilder(backendURL, true), httpClient.Do, forwardResponse)
}

func newDocumentsHandler(backendURL string, redis redis.Conn) http.Handler {
	cache := rediscache.NewWithClient(redis)
	cachingTransport := httpcache.NewTransport(cache)
	httpClient := &http.Client{Transport: cachingTransport}

	return newProxy(newRequestBuilder(backendURL, false), httpClient.Do, forwardResponse)
}
