package main

import (
	"log"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	address "github.com/example/ms-ecommerce/internal/services/address"
	"github.com/gin-gonic/gin"
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
	r.Use(middleware.GinLogging())
	r.Use(middleware.GinRecover())
	address.RegisterRoutes(r, dbConn)
	port := getenv("ADDRESS_PORT", "8083")
	addr := ":" + port
	log.Printf("address service running on %s", addr)
	log.Fatal(r.Run(addr))
}

func getenv(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}
