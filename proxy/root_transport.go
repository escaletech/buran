package proxy

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const cacheDuration = 7 * 24 * 60 * 60 // 7 days

type bodyTransformation func(body io.ReadCloser, req *http.Request) (io.ReadCloser, error)

func newRootTransport(inner http.RoundTripper, backendURL string) *rootTransport {
	return &rootTransport{inner, hostReplacer(backendURL)}
}

type rootTransport struct {
	transport     http.RoundTripper
	transformBody bodyTransformation
}

func (t *rootTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 200 && res.StatusCode < 400 {
		res.Header.Set("Cache-Control", "max-age="+strconv.Itoa(cacheDuration))
	}

	body, err := t.transformBody(res.Body, req)
	if err != nil {
		return nil, err
	}

	res.Body = body

	return res, nil
}

func hostReplacer(backendURL string) bodyTransformation {
	return func(body io.ReadCloser, req *http.Request) (io.ReadCloser, error) {
		content, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}

		proxyHost := req.URL.Query().Get(HostParamKey)

		fixed := strings.Replace(string(content), backendURL, proxyHost, -1)
		return ioutil.NopCloser(bytes.NewReader([]byte(fixed))), nil
	}
}
