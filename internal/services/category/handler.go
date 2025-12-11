package category

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)

	// Public: list and get
	r.HandleFunc("/api/v1/categories", makeListHandler(uc)).Methods("GET")
	r.HandleFunc("/api/v1/categories/{id}", makeGetHandler(uc)).Methods("GET")

	// Admin-only management
	r.Handle("/api/v1/categories", middleware.JWTAuth(middleware.RequireRole("admin")(makeCreateHandler(uc)))).Methods("POST")
	r.Handle("/api/v1/categories/{id}", middleware.JWTAuth(middleware.RequireRole("admin")(makeUpdateHandler(uc)))).Methods("PUT")
	r.Handle("/api/v1/categories/{id}", middleware.JWTAuth(middleware.RequireRole("admin")(makeDeleteHandler(uc)))).Methods("DELETE")
}

func makeCreateHandler(uc Usecase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		id, err := uc.Create(req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	})
}

func makeUpdateHandler(uc Usecase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		if err := uc.Update(id, req.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func makeDeleteHandler(uc Usecase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)
		if err := uc.Delete(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func makeListHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := uc.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": data})
	}
}

func makeGetHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, _ := strconv.ParseInt(vars["id"], 10, 64)
		c, err := uc.Get(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if c == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(c)
	}
}
