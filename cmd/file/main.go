package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	file "github.com/example/ms-ecommerce/internal/services/file"
	"github.com/gorilla/mux"
)

func main() {
	os.Setenv("DB_HOST", getenv("DB_HOST", "127.0.0.1"))
	dbConn, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	r := mux.NewRouter()
	r.Use(middleware.Logging)
	r.Use(middleware.Recover)
	file.RegisterRoutes(r, dbConn)
	port := getenv("FILE_PORT", "8085")
	addr := ":" + port
	log.Printf("file service running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
