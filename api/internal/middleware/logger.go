package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	Reset    = "\033[0m"
	Bold     = "\033[1m"
	Green    = "\033[32m"
	Yellow   = "\033[33m"
	Blue     = "\033[34m"
	Magenta  = "\033[35m"
	Cyan     = "\033[36m"
	White    = "\033[37m"
	BgGreen  = "\033[42m"
	BgRed    = "\033[41m"
	BgYellow = "\033[43m"
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
			logWebSocketConnection(r, start)
			next.ServeHTTP(w, r)
			return
		}

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		logHTTPRequest(rw, r, start)
	})
}

func logWebSocketConnection(r *http.Request, _ time.Time) {
	logEntry := fmt.Sprintf("%s┌ %s%s\n",
		White,
		time.Now().Format("2006-01-02 15:04:05"),
		Reset,
	)

	logEntry += fmt.Sprintf("%s├ %s%s %s%s %s%s%s\n",
		White,
		Cyan,
		padMethod("WS"),
		Blue,
		r.URL.Path,
		Green,
		"UPGRADE",
		Reset,
	)

	logEntry += fmt.Sprintf("%s└ %s%s%s · %s%s%s\n",
		White,
		Cyan,
		"WebSocket Connection",
		White,
		Yellow,
		r.RemoteAddr,
		Reset,
	)

	fmt.Print(logEntry)
}

func logHTTPRequest(rw *responseWriter, r *http.Request, start time.Time) {
	duration := time.Since(start)

	var durationStr string
	if duration.Milliseconds() < 1 {
		durationStr = fmt.Sprintf("%d µs", duration.Microseconds())
	} else if duration.Seconds() < 1 {
		durationStr = fmt.Sprintf("%d ms", duration.Milliseconds())
	} else {
		durationStr = fmt.Sprintf("%.2f s", duration.Seconds())
	}

	statusColor := getStatusColor(rw.statusCode)
	methodColor := getMethodColor(r.Method)

	logEntry := fmt.Sprintf("%s┌ %s%s\n",
		White,
		time.Now().Format("2006-01-02 15:04:05"),
		Reset,
	)

	logEntry += fmt.Sprintf("%s├ %s%s %s%s %s%d%s\n",
		White,
		methodColor,
		padMethod(r.Method),
		Blue,
		r.URL.Path,
		statusColor,
		rw.statusCode,
		Reset,
	)

	logEntry += fmt.Sprintf("%s└ %s%s%s · %s%d bytes%s\n",
		White,
		Cyan,
		durationStr,
		White,
		Yellow,
		rw.length,
		Reset,
	)

	fmt.Print(logEntry)
}

func getStatusColor(code int) string {
	switch {
	case code >= 500:
		return BgRed + White
	case code >= 400:
		return Yellow
	case code >= 300:
		return Magenta
	case code >= 200:
		return Green
	default:
		return White
	}
}

func getMethodColor(method string) string {
	switch method {
	case http.MethodGet:
		return Green
	case http.MethodPost:
		return Yellow
	case http.MethodPut:
		return Blue
	case http.MethodDelete:
		return BgRed + White
	case http.MethodPatch:
		return Magenta
	default:
		return White
	}
}

func padMethod(method string) string {
	return fmt.Sprintf("%-7s", method)
}

func isWebSocketUpgrade(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket" &&
		r.Header.Get("Connection") == "Upgrade"
}
