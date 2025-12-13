package jwtpkg

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getenv("JWT_SECRET", "supersecretjwtkey"))

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func GenerateToken(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken verifies the token and returns userID and role
func ParseToken(tokenStr string) (int64, string, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return 0, "", err
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		// extract user_id and role
		var uid int64
		switch v := claims["user_id"].(type) {
		case float64:
			uid = int64(v)
		case int64:
			uid = v
		}
		role, _ := claims["role"].(string)
		return uid, role, nil
	}
	return 0, "", errors.New("invalid token")
}

// GenerateResetToken creates a short-lived token for password reset
func GenerateResetToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"purpose": "reset",
		"exp":     time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseResetToken verifies reset token and returns userID if purpose matches
func ParseResetToken(tokenStr string) (int64, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := t.Claims.(jwt.MapClaims); ok && t.Valid {
		if purpose, _ := claims["purpose"].(string); purpose != "reset" {
			return 0, errors.New("invalid token purpose")
		}
		var uid int64
		switch v := claims["user_id"].(type) {
		case float64:
			uid = int64(v)
		case int64:
			uid = v
		}
		return uid, nil
	}
	return 0, errors.New("invalid token")
}
