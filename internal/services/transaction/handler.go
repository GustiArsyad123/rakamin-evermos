package transaction

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"database/sql"

	"github.com/example/ms-ecommerce/internal/pkg/middleware"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, dbConn *sql.DB) {
	repo := NewRepo(dbConn)
	uc := NewUsecase(repo, dbConn)
	// transactions require auth
	r.Handle("/api/v1/transactions", middleware.JWTAuth(makeCreateHandler(uc))).Methods("POST")
	r.Handle("/api/v1/transactions", middleware.JWTAuth(makeListHandler(uc))).Methods("GET")
	r.Handle("/api/v1/transactions/{id}", middleware.JWTAuth(makeGetHandler(uc))).Methods("GET")
	r.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
}

func makeCreateHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := middleware.GetUserID(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var req struct {
			AddressID int64     `json:"address_id"`
			Items     []ItemReq `json:"items"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid", http.StatusBadRequest)
			return
		}
		if req.AddressID == 0 {
			http.Error(w, "address_id is required", http.StatusBadRequest)
			return
		}
		id, err := uc.Create(uid, req.AddressID, req.Items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	}
}

func makeListHandler(uc Usecase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := middleware.GetUserID(r)
		log.Printf("Transaction list handler: uid=%v, ok=%v", uid, ok)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("ok"))
		return
		role, _ := middleware.GetRole(r)
		q := r.URL.Query()
		filters := map[string]string{}
		if v := q.Get("status"); v != "" {
			filters["status"] = v
		}
		if v := q.Get("store_id"); v != "" {
			filters["store_id"] = v
		}
		if v := q.Get("min_total"); v != "" {
			filters["min_total"] = v
		}
		if v := q.Get("max_total"); v != "" {
			filters["max_total"] = v
		}

		page := 1
		limit := 10
		if v := q.Get("page"); v != "" {
			if pi, err := strconv.Atoi(v); err == nil {
				page = pi
			}
		}
		if v := q.Get("limit"); v != "" {
			if li, err := strconv.Atoi(v); err == nil {
				limit = li
			}
		}
		data, total, err := uc.List(uid, role, filters, page, limit)
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
		role, _ := middleware.GetRole(r)
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, _ := strconv.ParseInt(idStr, 10, 64)
		t, logs, err := uc.Get(uid, id, role)
		if err != nil {
			if err.Error() == "forbidden" {
				http.Error(w, err.Error(), http.StatusForbidden)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if t == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"transaction": t, "logs": logs})
	}
}
