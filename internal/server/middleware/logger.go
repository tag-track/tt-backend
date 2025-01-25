package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs request details including method, path, status code, and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &responseWriterWrapper{w: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		duration := time.Since(start)
		log.Printf(
			"status=%d method=%s duration=%s path=%s query=%s",
			wrappedWriter.statusCode,
			r.Method,
			duration,
			r.URL.Path,
			r.URL.RawQuery,
		)
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture the status code.
type responseWriterWrapper struct {
	w           http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
	}
	rw.w.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.w.Write(b)
}

func (rw *responseWriterWrapper) Header() http.Header {
	return rw.w.Header()
}
