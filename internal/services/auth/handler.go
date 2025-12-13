package auth

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"time"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	r.HandleFunc("/api/v1/auth/register", makeRegisterHandler(uc)).Methods("POST")
	log.Printf("registered POST /api/v1/auth/register")
	r.HandleFunc("/api/v1/auth/login", makeLoginHandler(uc)).Methods("POST")
	log.Printf("registered POST /api/v1/auth/login")
	r.HandleFunc("/api/v1/auth/forgot-password", makeForgotHandler(uc)).Methods("POST")
	log.Printf("registered POST /api/v1/auth/forgot-password")
	r.HandleFunc("/api/v1/auth/reset-password", makeResetHandler(uc)).Methods("POST")
	log.Printf("registered POST /api/v1/auth/reset-password")
	r.HandleFunc("/api/v1/auth/sso/google", makeSSOHandler(uc)).Methods("POST")
	log.Printf("registered POST /api/v1/auth/sso/google")
	// Admin-only list users (register both trailing and non-trailing variants)
	usersHandler := middleware.JWTAuth(middleware.RequireRole("admin")(makeListUsersHandler(uc)))
	r.Handle("/api/v1/auth/users", usersHandler).Methods("GET")
	r.Handle("/api/v1/auth/users/", usersHandler).Methods("GET")
	log.Printf("registered GET /api/v1/auth/users and /api/v1/auth/users/")
	// Get user by id: requires JWT; allowed if admin or owner
	// Keep this under the auth prefix for consistency with other auth routes
	r.Handle("/api/v1/auth/users/{id}", middleware.JWTAuth(makeGetUserHandler(uc))).Methods("GET")
	log.Printf("registered GET /api/v1/auth/users/{id}")
	// Update user (owner or admin). Admin may change role.
	r.Handle("/api/v1/auth/users/{id}", middleware.JWTAuth(makeUpdateUserHandler(uc))).Methods("PUT")
	log.Printf("registered PUT /api/v1/auth/users/{id}")
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
		if req.Name == "" || req.Email == "" || req.Phone == "" || req.Password == "" {
			http.Error(w, "name, email, phone, and password are required", http.StatusBadRequest)
			return
		}
		u := &models.User{Name: req.Name, Email: req.Email, Phone: req.Phone, Password: req.Password}
		token, userID, err := uc.Register(u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		refresh, refreshExp, err := uc.IssueRefreshToken(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":              token,
			"expires_in":         3600,
			"expires_at":         expiresAt.Format(time.RFC3339),
			"refresh_token":      refresh,
			"refresh_expires_at": refreshExp.Format(time.RFC3339),
		})
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
		json.NewEncoder(w).Encode(map[string]interface{}{"reset_token": token, "expires_in": 3600})
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
		token, userID, err := uc.SSOLogin(req.IDToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		refresh, refreshExp, err := uc.IssueRefreshToken(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":              token,
			"expires_in":         3600,
			"expires_at":         expiresAt.Format(time.RFC3339),
			"refresh_token":      refresh,
			"refresh_expires_at": refreshExp.Format(time.RFC3339),
		})
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
		token, userID, err := uc.Login(req.Email, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		refresh, refreshExp, err := uc.IssueRefreshToken(userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token":              token,
			"expires_in":         3600,
			"expires_at":         expiresAt.Format(time.RFC3339),
			"refresh_token":      refresh,
			"refresh_expires_at": refreshExp.Format(time.RFC3339),
		})
	}
}

func makeListUsersHandler(uc Usecase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// parse pagination
		q := r.URL.Query()
		page := 1
		limit := 10
		// enforce a reasonable max limit
		const maxLimit = 100
		if p := q.Get("page"); p != "" {
			if v, err := strconv.Atoi(p); err == nil && v > 0 {
				page = v
			}
		}
		if l := q.Get("limit"); l != "" {
			if v, err := strconv.Atoi(l); err == nil && v > 0 {
				limit = v
			}
		}
		if limit > maxLimit {
			limit = maxLimit
		}
		search := q.Get("search")
		users, total, err := uc.ListUsers(page, limit, search)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp := map[string]interface{}{
			"data": users,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
}

func makeGetUserHandler(uc Usecase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		// permission: admin or owner
		reqUID, ok := middleware.GetUserID(r)
		role, _ := middleware.GetRole(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if role != "admin" && reqUID != id {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		u, err := uc.GetUserByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if u == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(u)
	})
}

func makeUpdateUserHandler(uc Usecase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		reqUID, ok := middleware.GetUserID(r)
		role, _ := middleware.GetRole(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var body struct {
			Name  *string `json:"name"`
			Phone *string `json:"phone"`
			Role  *string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		// enforce authorization in usecase
		if err := uc.UpdateUser(reqUID, id, role, body.Name, body.Phone, body.Role); err != nil {
			if err.Error() == "forbidden" {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
