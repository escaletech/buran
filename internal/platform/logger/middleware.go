package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func NewMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return &middleware{next}
	}
}

type middleware struct {
	next http.Handler
}

func (h *middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	httpFields := logrus.Fields{
		"method": r.Method,
		"url":    r.URL.String(),
	}

	fields := logrus.Fields{
		"http": httpFields,
	}

	lw := newLoggerReponseWriter(w)
	h.next.ServeHTTP(lw, r)

	latency := time.Since(start)
	message := fmt.Sprintf("%v %v | %v | %v \"%v\"", r.Method, r.RequestURI, latency, lw.status, http.StatusText(lw.status))

	fields["duration"] = latency / time.Millisecond
	fields["from_cache"] = lw.Header().Get("X-From-Cache")
	httpFields["status_code"] = lw.status

	if lw.status >= http.StatusBadRequest && lw.status < http.StatusInternalServerError {
		logger.WithFields(fields).Warn(message)
		return
	}

	if lw.status >= http.StatusInternalServerError {
		logger.WithFields(fields).Error(message)
		return
	}

	logger.WithFields(fields).Info(message)
}

// loggerReponseWriter - wrapper to ResponseWriter
type loggerReponseWriter struct {
	http.Flusher
	http.ResponseWriter
	http.CloseNotifier
	status int
}

func newLoggerReponseWriter(w http.ResponseWriter) *loggerReponseWriter {
	var flusher http.Flusher
	var cNotifier http.CloseNotifier
	var ok bool
	if flusher, ok = w.(http.Flusher); !ok {
		flusher = nil
	}

	if cNotifier, ok = w.(http.CloseNotifier); !ok {
		cNotifier = nil
	}

	return &loggerReponseWriter{flusher, w, cNotifier, http.StatusOK}
}

func (lrw *loggerReponseWriter) Write(body []byte) (int, error) {
	return lrw.ResponseWriter.Write(body)
}

func (lrw *loggerReponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}
