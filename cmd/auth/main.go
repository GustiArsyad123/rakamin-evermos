package main

import (
	"log"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	auth "github.com/example/ms-ecommerce/internal/services/auth"
	category "github.com/example/ms-ecommerce/internal/services/category"
	"github.com/gin-gonic/gin"
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
	r := gin.New()
	// attach middleware for logging and recovery to help with debugging
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.GinRateLimit())
	r.Use(middleware.GinMetricsMiddleware("auth"))
	auth.RegisterRoutes(r, dbConn)
	category.RegisterRoutes(r, dbConn)

	// Add metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	port := getenv("AUTH_PORT", "8080")
	addr := ":" + port
	log.Printf("auth service running on %s", addr)
	log.Fatal(r.Run(addr))
}

func getenv(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}
