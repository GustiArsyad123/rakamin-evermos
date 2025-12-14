package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	jwtpkg "github.com/example/ms-ecommerce/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type ctxKey string

const (
	ctxUserID ctxKey = "user_id"
	ctxRole   ctxKey = "role"
	ctxToken  ctxKey = "token"
)

// RateLimit adalah middleware untuk membatasi rate request per IP
var limiter = rate.NewLimiter(10, 20) // 10 requests/second, burst 20

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GinRateLimit adalah middleware Gin untuk membatasi rate request per IP
func GinRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// JWTAuth adalah middleware untuk memverifikasi token JWT dari header Authorization.
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := extractToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userID, role, err := jwtpkg.ParseToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserID, userID)
		ctx = context.WithValue(ctx, ctxRole, role)
		ctx = context.WithValue(ctx, ctxToken, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractToken(r *http.Request) (string, error) {
	log.Printf("AUTH HEADER: %q", r.Header.Get("Authorization"))
	log.Printf("COOKIE access_token: %+v", r.Cookies())

	// 1️⃣ Try Authorization Header
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	log.Printf("AUTH len=%d, prefix=%q", len(auth), func() string {
		if len(auth) >= 10 {
			return auth[:10]
		}
		return auth
	}())

	if auth := strings.TrimSpace(r.Header.Get("Authorization")); auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", errors.New("invalid authorization scheme")
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			return "", errors.New("missing token")
		}

		return token, nil
	}

	// 2️⃣ Fallback to Cookie
	cookie, err := r.Cookie("access_token")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", errors.New("missing authorization header or cookie")
		}
		return "", err
	}

	if cookie.Value == "" {
		return "", errors.New("empty token in cookie")
	}

	return cookie.Value, nil
}

// RequireRole checks that the injected role matches required
func RequireRole(required string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Context().Value(ctxRole)
			if v == nil || v.(string) != required {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID extracts user id from context
func GetUserID(r *http.Request) (int64, bool) {
	v := r.Context().Value(ctxUserID)
	if v == nil {
		return 0, false
	}
	uid, ok := v.(int64)
	return uid, ok
}

// GetRole extracts role from context
func GetRole(r *http.Request) (string, bool) {
	v := r.Context().Value(ctxRole)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// Gin-compatible middleware functions

// GinJWTAuth middleware for Gin framework
func GinJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractTokenGin(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		userID, role, err := jwtpkg.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set values in Gin context
		c.Set(ctxUserID, userID)
		c.Set(ctxRole, role)
		c.Set(ctxToken, token)

		c.Next()
	}
}

// GinRequireRole checks that the injected role matches required for Gin
func GinRequireRole(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, exists := c.Get(ctxRole)
		if !exists || v.(string) != required {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// extractTokenGin extracts token from Gin context
func extractTokenGin(c *gin.Context) (string, error) {
	// 1️⃣ Try Authorization Header
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", errors.New("invalid authorization scheme")
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			return "", errors.New("missing token")
		}

		return token, nil
	}

	// 2️⃣ Fallback to Cookie
	cookie, err := c.Cookie("access_token")
	if err != nil {
		return "", errors.New("missing authorization header or cookie")
	}

	if cookie == "" {
		return "", errors.New("empty token in cookie")
	}

	return cookie, nil
}

// GinGetUserID extracts user id from Gin context
func GinGetUserID(c *gin.Context) (int64, bool) {
	v, exists := c.Get(ctxUserID)
	if !exists {
		return 0, false
	}
	uid, ok := v.(int64)
	return uid, ok
}

// GinGetRole extracts role from Gin context
func GinGetRole(c *gin.Context) (string, bool) {
	v, exists := c.Get(ctxRole)
	if !exists {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
