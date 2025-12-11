package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	jwtpkg "github.com/example/ms-ecommerce/internal/pkg/jwt"
)

type ctxKey string

const (
	ctxUserID ctxKey = "user_id"
	ctxRole   ctxKey = "role"
)

// JWTAuth parses Authorization header and injects user id and role into context
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized: missing or malformed authorization header", http.StatusUnauthorized)
			return
		}
		tok := strings.TrimPrefix(auth, "Bearer ")
		uid, role, err := jwtpkg.ParseToken(tok)
		if err != nil {
			log.Printf("JWTAuth: failed to parse token: %v", err) // Add logging for easier debugging
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserID, uid)
		ctx = context.WithValue(ctx, ctxRole, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
