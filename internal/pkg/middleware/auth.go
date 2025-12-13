package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	jwtpkg "github.com/example/ms-ecommerce/internal/pkg/jwt"
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
