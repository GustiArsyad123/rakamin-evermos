package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	r.HandleFunc("/api/v1/auth/register", makeRegisterHandler(uc)).Methods("POST")
	r.HandleFunc("/api/v1/auth/login", makeLoginHandler(uc)).Methods("POST")
	r.HandleFunc("/api/v1/auth/forgot-password", makeForgotHandler(uc)).Methods("POST")
	r.HandleFunc("/api/v1/auth/reset-password", makeResetHandler(uc)).Methods("POST")
	r.HandleFunc("/api/v1/auth/sso/google", makeSSOHandler(uc)).Methods("POST")
}

func makeRegisterHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Phone    string `json:"phone"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		u := &models.User{Name: req.Name, Email: req.Email, Phone: req.Phone, Password: req.Password}
		token, err := uc.Register(u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

func makeForgotHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		token, err := uc.ForgotPassword(req.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// In real app send email; for now return token so it can be used in tests/Postman
		json.NewEncoder(w).Encode(map[string]string{"reset_token": token})
	}
}

func makeResetHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" || req.Password == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if err := uc.ResetPassword(req.Token, req.Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func makeSSOHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			IDToken string `json:"id_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		token, err := uc.SSOLogin(req.IDToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

func makeLoginHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		token, err := uc.Login(req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}
