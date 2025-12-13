package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	auth "github.com/example/ms-ecommerce/internal/services/auth"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// allow overriding DB via env
	os.Setenv("DB_HOST", getenv("DB_HOST", "127.0.0.1"))
	dbConn, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	// Ensure minimal auth-related tables exist (avoids startup failure when
	// the DB was not initialized from `sql/` files).
	if err := db.EnsureAuthTables(dbConn); err != nil {
		log.Fatalf("ensure auth tables: %v", err)
	}
	r := mux.NewRouter()
	// attach middleware for logging and recovery to help with debugging
	r.Use(middleware.Logging)
	r.Use(middleware.Recover)
	r.Use(middleware.RateLimit)
	r.Use(middleware.MetricsMiddleware("auth"))
	auth.RegisterRoutes(r, dbConn)

	// Add metrics endpoint
	r.Path("/metrics").Handler(promhttp.Handler())
	port := getenv("AUTH_PORT", "8080")
	addr := ":" + port
	log.Printf("auth service running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getenv(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}
