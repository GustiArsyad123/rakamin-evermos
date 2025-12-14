package main

import (
	"log"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/cache"
	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	product "github.com/example/ms-ecommerce/internal/services/product"
	"github.com/gin-gonic/gin"
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

	r := gin.New()
	r.Use(middleware.GinLogging())
	r.Use(middleware.GinRecover())
	r.Use(middleware.GinRateLimit())
	r.Use(middleware.GinMetricsMiddleware("product"))
	product.RegisterRoutes(r, dbConn, productCache)

	// Add metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	port := getenv("PRODUCT_PORT", "8081")
	addr := ":" + port
	log.Printf("product service running on %s", addr)
	log.Fatal(r.Run(addr))
}

func getenv(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}
