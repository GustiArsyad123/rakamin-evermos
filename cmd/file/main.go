package main

import (
	"log"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	file "github.com/example/ms-ecommerce/internal/services/file"
	"github.com/gin-gonic/gin"
)

func main() {
	os.Setenv("DB_HOST", getenv("DB_HOST", "127.0.0.1"))
	dbConn, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	r := gin.New()
	r.Use(middleware.GinLogging())
	r.Use(middleware.GinRecover())
	file.RegisterRoutes(r, dbConn)
	port := getenv("FILE_PORT", "8085")
	addr := ":" + port
	log.Printf("file service running on %s", addr)
	log.Fatal(r.Run(addr))
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
