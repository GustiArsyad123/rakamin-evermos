package address

import (
	"encoding/json"
	"net/http"
	"strconv"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/example/ms-ecommerce/internal/pkg/models"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo)
	r.Handle("/api/v1/addresses", middleware.JWTAuth(makeCreateHandler(uc))).Methods("POST")
	r.Handle("/api/v1/addresses", middleware.JWTAuth(makeListHandler(uc))).Methods("GET")
	r.Handle("/api/v1/addresses/{id}", middleware.JWTAuth(makeGetHandler(uc))).Methods("GET")
	r.Handle("/api/v1/addresses/{id}", middleware.JWTAuth(makeUpdateHandler(uc))).Methods("PUT")
	r.Handle("/api/v1/addresses/{id}", middleware.JWTAuth(makeDeleteHandler(uc))).Methods("DELETE")
}

func makeCreateHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var req struct {
			Label      string `json:"label"`
			Address    string `json:"address"`
			City       string `json:"city"`
			PostalCode string `json:"postal_code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Address == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		a := &models.Address{Label: req.Label, Address: req.Address, City: req.City, PostalCode: req.PostalCode}
		id, err := uc.CreateAddress(uid, a)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

func makeListHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		q := r.URL.Query()
		filters := map[string]string{}
		if v := q.Get("label"); v != "" {
			filters["label"] = v
		}
		if v := q.Get("city"); v != "" {
			filters["city"] = v
		}
		if v := q.Get("postal_code"); v != "" {
			filters["postal_code"] = v
		}

		page := 1
		limit := 10
		if v := q.Get("page"); v != "" {
			if pi, err := strconv.Atoi(v); err == nil && pi > 0 {
				page = pi
			}
		}
		if v := q.Get("limit"); v != "" {
			if li, err := strconv.Atoi(v); err == nil && li > 0 {
				limit = li
			}
		}

		data, total, err := uc.ListAddresses(uid, filters, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": data, "pagination": map[string]interface{}{"page": page, "limit": limit, "total": total}})
	}
}

func makeGetHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		a, err := uc.GetAddress(uid, id)
		if err != nil {
			if err.Error() == "forbidden" {
				http.Error(w, err.Error(), http.StatusForbidden)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if a == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(a)
	}
}

func makeUpdateHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var req struct {
			Label      string `json:"label"`
			Address    string `json:"address"`
			City       string `json:"city"`
			PostalCode string `json:"postal_code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Address == "" {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		err = uc.UpdateAddress(uid, id, req.Label, req.Address, req.City, req.PostalCode)
		if err != nil {
			if err.Error() == "forbidden" {
				http.Error(w, err.Error(), http.StatusForbidden)
			} else if err.Error() == "not found" {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func makeDeleteHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		err = uc.DeleteAddress(uid, id)
		if err != nil {
			if err.Error() == "forbidden" {
				http.Error(w, err.Error(), http.StatusForbidden)
			} else if err.Error() == "not found" {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
