package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func Direct(backendURL string) *httputil.ReverseProxy {
	uri, err := url.Parse(backendURL)
	if err != nil {
		panic(err)
	}

	host := strings.Replace(uri.Host, "cdn.prismic.io", "prismic.io", -1)
	scheme := uri.Scheme

	director := func(req *http.Request) {
		req.Header.Set("X-Forwarded-Host", req.Host)

		req.Host = uri.Host
		req.URL.Host = host
		req.URL.Scheme = scheme
	}
	return &httputil.ReverseProxy{Director: director}
}
