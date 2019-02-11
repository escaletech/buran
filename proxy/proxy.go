package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/escaleseo/prismic-proxy-cache/logger"
)

const HostParamKey = "proxy-host"

var log = logger.Get()

var rootHeaderBlacklist = map[string]struct{}{
	"Cache-Control":   struct{}{},
	"Pragma":          struct{}{},
	"Accept-Encoding": struct{}{},
}

type requestBuilder func(r *http.Request) (*http.Request, error)

type responseForwarder func(w http.ResponseWriter, res *http.Response) error

type httpRequester func(*http.Request) (*http.Response, error)

func newProxy(buildRequest requestBuilder, do httpRequester, forward responseForwarder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := buildRequest(r)
		if err != nil {
			serverError(w, err)
			return
		}

		res, err := do(req)
		if err != nil {
			serverError(w, err)
			return
		}

		if err := forward(w, res); err != nil {
			log.WithError(err).Fatal(err)
		}
	}
}

func forwardResponse(w http.ResponseWriter, res *http.Response) error {
	header := w.Header()
	for k, v := range res.Header {
		header[k] = v
	}

	w.WriteHeader(res.StatusCode)

	if res.Body != nil {
		defer res.Body.Close()
		if _, err := io.Copy(w, res.Body); err != nil {
			return err
		}
	}

	return nil
}

func serverError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func newRequestBuilder(backendURL string, isAPIRoot bool) requestBuilder {
	return func(r *http.Request) (*http.Request, error) {
		targetURL := strings.TrimSuffix(backendURL+r.URL.Path, "/")
		if isAPIRoot {
			targetURL += hostVariationParam(r)
		} else if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		req, err := http.NewRequest("GET", targetURL, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create target request")
		}

		if isAPIRoot {
			for k, v := range r.Header {
				if _, blacklisted := rootHeaderBlacklist[k]; !blacklisted {
					req.Header[k] = v
				}
			}
		} else {
			req.Header = r.Header
		}

		return req, nil
	}
}

func hostVariationParam(r *http.Request) string {
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}
	return fmt.Sprintf("?%v=%v", HostParamKey, url.QueryEscape(proto+"://"+r.Host))
}
