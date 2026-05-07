package middleware

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	length     int64
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.length += int64(n)
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if isWebSocketUpgrade(r) {
			logWebSocketConnection(r)
			next.ServeHTTP(w, r)
			return
		}

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		logHTTPRequest(rw, r, start)
	})
}

func logWebSocketConnection(r *http.Request) {
	logrus.WithFields(logrus.Fields{
		"req": map[string]interface{}{
			"method":        "WS",
			"url":           r.URL.Path,
			"remoteAddress": r.RemoteAddr,
		},
	}).Info("websocket upgrade")
}

func logHTTPRequest(rw *responseWriter, r *http.Request, start time.Time) {
	duration := time.Since(start)
	logrus.WithFields(logrus.Fields{
		"req": map[string]interface{}{
			"method":        r.Method,
			"url":           r.URL.Path,
			"remoteAddress": r.RemoteAddr,
		},
		"res": map[string]interface{}{
			"statusCode": rw.statusCode,
		},
		"responseTime": duration.Milliseconds(),
		"bytes":        rw.length,
	}).Info("request completed")
}

func isWebSocketUpgrade(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket" &&
		r.Header.Get("Connection") == "Upgrade"
}
