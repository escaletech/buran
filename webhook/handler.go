package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/escaleseo/prismic-proxy-cache/cache"
	"github.com/escaleseo/prismic-proxy-cache/logger"
)

var log = logger.Get()

type invalidator func() error

func New(cache cache.Provider) http.Handler {
	return &webhookHandler{cache.Invalidate}
}

type webhookHandler struct {
	invalidate invalidator
}

func (h *webhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Type string `json:"type,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if body.Type != "api-update" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("unknown webhook type " + body.Type))
		return
	}

	if err := h.invalidate(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
