package main

import (
	"log"
	"net/http"
	"os"

	"github.com/example/ms-ecommerce/internal/pkg/db"
	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	auth "github.com/example/ms-ecommerce/internal/services/auth"
	"github.com/gorilla/mux"
)

func main() {
	// DB connection details are now read from environment variables by the driver
	dbConn, err := db.NewMySQL()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer dbConn.Close()

	r := mux.NewRouter()
	// attach middleware for logging and recovery to help with debugging
	r.Use(middleware.Logging)
	r.Use(middleware.Recover)

	// Initialize layers
	repo := auth.NewRepository(dbConn)
	usecase := auth.NewUsecase(repo)
	auth.RegisterRoutes(r, usecase) // Pass the usecase to RegisterRoutes

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
