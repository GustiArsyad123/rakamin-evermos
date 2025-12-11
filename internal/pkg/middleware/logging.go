package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"
)

// responseWriter captures status and size for logging
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// isDebug checks the DEBUG env flag
func isDebug() bool {
	return os.Getenv("DEBUG") == "true"
}

// Logging is a middleware that logs method, path, status and duration.
// When DEBUG=true it also logs a truncated request body (up to maxBodyLength).
func Logging(next http.Handler) http.Handler {
	const maxBodyLength = 10 * 1024 // 10KB
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var bodySnippet string
		if isDebug() && r.Body != nil {
			// read body but preserve it for next handler
			buf, err := io.ReadAll(io.LimitReader(r.Body, int64(maxBodyLength)+1))
			if err == nil {
				if len(buf) > maxBodyLength {
					bodySnippet = string(buf[:maxBodyLength]) + "...[truncated]"
				} else {
					bodySnippet = string(buf)
				}
			}
			r.Body = io.NopCloser(bytes.NewBuffer(buf))
		}

		// redact Authorization header for logging
		auth := r.Header.Get("Authorization")
		if auth != "" {
			r.Header.Set("Authorization", "REDACTED")
		}

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)
		dur := time.Since(start)
		if rw.status == 0 {
			rw.status = http.StatusOK
		}

		if isDebug() && bodySnippet != "" {
			log.Printf("%s %s %d %dB %s body=%q", r.Method, r.URL.Path, rw.status, rw.size, dur, bodySnippet)
		} else {
			log.Printf("%s %s %d %dB %s", r.Method, r.URL.Path, rw.status, rw.size, dur)
		}
	})
}

// Recover logs panic and returns 500. When DEBUG=true it includes a stack trace.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if isDebug() {
					log.Printf("panic: %v\n%s", rec, debug.Stack())
				} else {
					log.Printf("panic: %v", rec)
				}
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
