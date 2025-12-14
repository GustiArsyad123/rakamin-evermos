package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
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

		// Note: Authorization header is not logged for security

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

// Gin-compatible middleware functions

// GinLogging is a middleware that logs method, path, status and duration for Gin.
// When DEBUG=true it also logs a truncated request body (up to maxBodyLength).
func GinLogging() gin.HandlerFunc {
	const maxBodyLength = 10 * 1024 // 10KB
	return func(c *gin.Context) {
		var bodySnippet string
		if isDebug() && c.Request.Body != nil {
			// read body but preserve it for next handler
			buf, err := io.ReadAll(io.LimitReader(c.Request.Body, int64(maxBodyLength)+1))
			if err == nil {
				if len(buf) > maxBodyLength {
					bodySnippet = string(buf[:maxBodyLength]) + "...[truncated]"
				} else {
					bodySnippet = string(buf)
				}
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))
		}

		// Note: Authorization header is not logged for security

		start := time.Now()
		c.Next()
		dur := time.Since(start)

		status := c.Writer.Status()
		size := c.Writer.Size()

		if isDebug() && bodySnippet != "" {
			log.Printf("%s %s %d %dB %s body=%q", c.Request.Method, c.Request.URL.Path, status, size, dur, bodySnippet)
		} else {
			log.Printf("%s %s %d %dB %s", c.Request.Method, c.Request.URL.Path, status, size, dur)
		}
	}
}

// GinRecover logs panic and returns 500 for Gin. When DEBUG=true it includes a stack trace.
func GinRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				if isDebug() {
					log.Printf("panic: %v\n%s", rec, debug.Stack())
				} else {
					log.Printf("panic: %v", rec)
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				c.Abort()
			}
		}()
		c.Next()
	}
}
