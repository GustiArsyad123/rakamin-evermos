package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/cache"
	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	product "github.com/example/ms-ecommerce/internal/services/product"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	os.Setenv("DB_HOST", getenv("DB_HOST", "127.0.0.1"))
	dbConn, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	// Initialize Redis cache
	redisClient, err := db.NewRedis()
	if err != nil {
		log.Printf("redis connect failed, continuing without cache: %v", err)
		redisClient = nil
	}
	var productCache *cache.ProductCache
	if redisClient != nil {
		productCache = cache.NewProductCache(db.NewRedisCache(redisClient))
	}

	r := mux.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.Recover)
	r.Use(middleware.RateLimit)
	r.Use(middleware.MetricsMiddleware("product"))
	product.RegisterRoutes(r, dbConn, productCache)

	// Add metrics endpoint
	r.Path("/metrics").Handler(promhttp.Handler())
	port := getenv("PRODUCT_PORT", "8081")
	addr := ":" + port
	log.Printf("product service running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getenv(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}
