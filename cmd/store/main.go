package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	store "github.com/example/ms-ecommerce/internal/services/store"
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
	store.RegisterRoutes(r, dbConn)
	port := getenv("STORE_PORT", "8084")
	addr := ":" + port
	log.Printf("store service running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
