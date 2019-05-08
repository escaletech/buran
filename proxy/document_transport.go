package proxy

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

type bodyTransformation func(body io.ReadCloser, req *http.Request) (io.ReadCloser, error)

func newDocumentTransport(inner http.RoundTripper) *documentTransport {
	return &documentTransport{inner, replaceImagesURLProtocol()}
}

type documentTransport struct {
	transport     http.RoundTripper
	transformBody bodyTransformation
}

func (t *documentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	body, err := t.transformBody(res.Body, req)
	if err != nil {
		return nil, err
	}

	res.Body = body

	return res, nil
}

var regex = regexp.MustCompile(`(?m)"url":"http(:[^"]+)"`)

func replaceImagesURLProtocol() bodyTransformation {
	return func(body io.ReadCloser, req *http.Request) (io.ReadCloser, error) {
		content, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
		fixed := regex.ReplaceAllString(string(content), `"url":"https$1"`)

		return ioutil.NopCloser(bytes.NewReader([]byte(fixed))), nil
	}
}
