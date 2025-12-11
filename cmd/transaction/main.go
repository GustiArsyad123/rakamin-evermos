package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	txn "github.com/example/ms-ecommerce/internal/services/transaction"
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
	txn.RegisterRoutes(r, dbConn)
	port := getenv("TRANSACTION_PORT", "8082")
	addr := ":" + port
	log.Printf("transaction service running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func getenv(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}
