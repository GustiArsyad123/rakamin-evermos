package auth

import (
	"net/http"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router, usecase Usecase) {
	h := NewHandler(usecase)

	// Create a subrouter for /api/v1/auth
	subrouter := router.PathPrefix("/api/v1/auth").Subrouter()

	// Public routes on the subrouter
	subrouter.HandleFunc("/register", h.handleRegister).Methods("POST")
	subrouter.HandleFunc("/login", h.handleLogin).Methods("POST")
	subrouter.HandleFunc("/forgot-password", h.handleForgotPassword).Methods("POST")
	subrouter.HandleFunc("/reset-password", h.handleResetPassword).Methods("POST")
	subrouter.HandleFunc("/sso/google", h.handleSSOLogin).Methods("POST")

	// Protected routes on the subrouter
	subrouter.Handle("/users", middleware.JWTAuth(http.HandlerFunc(h.handleGetAllUsers))).Methods("GET")
	subrouter.Handle("/users/{id}", middleware.JWTAuth(http.HandlerFunc(h.handleGetUserByID))).Methods("GET")
	subrouter.Handle("/me", middleware.JWTAuth(http.HandlerFunc(h.handleGetMyProfile))).Methods("GET")
}
